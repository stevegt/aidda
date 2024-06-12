package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Action struct to represent an action with its arguments
type Action struct {
	Name string
	Args []string
}

// Function to create Docker client
func createDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// Function to execute action in Docker container
func executeActionInContainer(action Action) (string, error) {
	cli, err := createDockerClient()
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "golang:1.18",
		Cmd:   []string{"go", "run", "actionRunner.go", action.Name, strings.Join(action.Args, " ")},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	defer cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}
	defer out.Close()

	logs, err := ioutil.ReadAll(out)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// Function to execute actions and handle errors
func executeActions(actions []Action) ([]string, error) {
	var results []string
	for _, action := range actions {
		var result string
		var err error

		if action.Name == "queryUser" {
			result = handleUserQuery(action.Args)
		} else {
			result, err = executeActionInContainer(action)
		}

		if err != nil {
			results = append(results, fmt.Sprintf("%s: error: %s", action.Name, err.Error()))
		} else {
			results = append(results, fmt.Sprintf("%s: %s", action.Name, result))
		}
	}
	return results, nil
}

// Function to handle user queries directly
func handleUserQuery(args []string) string {
	return "Handled user query"
}

// Function to format the valid actions for GPT prompt
func formatValidActions(validActions map[string]string) string {
	var sb strings.Builder
	sb.WriteString("Valid actions:\n")
	for action, usage := range validActions {
		sb.WriteString(fmt.Sprintf("%s: %s\n", action, usage))
	}
	return sb.String()
}

// GPT request and response types
type GPTRequest struct {
	Prompt string `json:"prompt"`
}

type GPTResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// Function to query GPT-4o
func queryGPT(userInstruction string, validActions map[string]string) (string, error) {
	apiKey := os.Getenv("GPT_API_KEY")
	url := "https://api.openai.com/v1/engines/gpt-4o/completions"
	actionsStr := formatValidActions(validActions)
	prompt := fmt.Sprintf("User instruction: '%s'\n\n%s\nPlease provide the actions to be taken.", userInstruction, actionsStr)
	reqBody, _ := json.Marshal(GPTRequest{Prompt: prompt})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var gptResp GPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&gptResp); err != nil {
		return "", err
	}

	return gptResp.Choices[0].Text, nil
}

// Function to parse actions from GPT response
func parseActions(response string, validActions map[string]string) []Action {
	var actions []Action
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		for action := range validActions {
			if strings.HasPrefix(line, action) {
				parts := strings.SplitN(line, " ", 2)
				var args []string
				if len(parts) > 1 {
					args = strings.Split(parts[1], " ")
				}
				actions = append(actions, Action{Name: action, Args: args})
			}
		}
	}
	return actions
}

// Helper function to launch the editor with a template
func launchEditor(template string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default to vi if $EDITOR is not set
	}

	tmpfile, err := ioutil.TempFile("", "user_query_*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(template)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// Function to format the message template for the editor
func formatTemplate(gptMessage string) string {
	var sb strings.Builder
	sb.WriteString("# Please type your query below. Lines starting with '#' will be ignored.\n")
	sb.WriteString("# GPT Message:\n")
	lines := strings.Split(gptMessage, "\n")
	for _, line := range lines {
		sb.WriteString("# " + line + "\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

// Function to clean user input by removing commented lines
func cleanUserQuery(input string) string {
	var cleaned strings.Builder
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			cleaned.WriteString(line + "\n")
		}
	}
	return strings.TrimSpace(cleaned.String())
}

// Main function
func main() {
	// Example GPT message
	gptMessage := "Here is a message from GPT. Please provide your query based on this information."

	// Step 1: Launch editor with template
	template := formatTemplate(gptMessage)
	userQuery, err := launchEditor(template)
	if err != nil {
		log.Fatalf("Error launching editor: %v\n", err)
	}

	// Step 2: Clean up user input (remove comments)
	userQuery = cleanUserQuery(userQuery)

	// Define valid actions
	validActions := map[string]string{
		"queryGopls":         "queryGopls args...",
		"fetchLinesFromFile": "fetchLinesFromFile path startLine endLine",
		"fetchFile":          "fetchFile path",
		"changeFile":         "changeFile path newContent",
		"changeLines":        "changeLines path startLine endLine newContent",
		"createFile":         "createFile path content",
		"runTests":           "runTests packagePath",
		"queryUser":          "queryUser query",
	}

	// Step 3: Forward the query to GPT-4o with valid actions
	gptResponse, err := queryGPT(userQuery, validActions)
	if err != nil {
		log.Fatalf("Error querying GPT-4o: %v\n", err)
	}

	// Step 4: Parse actions from GPT-4o response
	actions := parseActions(gptResponse, validActions)

	// Step 5: Execute actions
	results, err := executeActions(actions)
	if err != nil {
		log.Fatalf("Error executing actions: %v\n", err)
	}

	// Step 6: Return results to GPT-4o for the next actions
	var gptNextPrompt strings.Builder
	for _, result := range results {
		gptNextPrompt.WriteString(fmt.Sprintf("%s\n", result))
	}
	gptNextResponse, err := queryGPT(gptNextPrompt.String(), validActions)
	if err != nil {
		log.Fatalf("Error querying GPT-4o: %v\n", err)
	}

	// Process next actions as needed
	fmt.Println("Next actions from GPT-4o:", gptNextResponse)
}
