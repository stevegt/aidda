package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// GPTRequest struct represents the request body for the OpenAI API.
type GPTRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens"`
	Temperature float64  `json:"temperature"`
	TopP        float64  `json:"top_p"`
	N           int      `json:"n"`
	Stop        []string `json:"stop,omitempty"`
}

// GPTResponse struct represents the response from the OpenAI API.
type GPTResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// Function to query GPT-4 API.
func queryGPT(userInstruction string, validActions map[string]string) (string, error) {
	apiKey := os.Getenv("GPT_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("API key not set")
	}

	actionsStr := formatValidActions(validActions)
	prompt := fmt.Sprintf("User instruction: '%s'\n\n%s\nPlease provide the actions to be taken.", userInstruction, actionsStr)

	requestBody := GPTRequest{
		Model:       "gpt-4",
		Prompt:      prompt,
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        1.0,
		N:           1,
		Stop:        []string{"\n"},
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("error: %s", string(bodyBytes))
	}

	var gptResp GPTResponse
	if err := json.NewDecoder(resp.Body).Decode(&gptResp); err != nil {
		return "", err
	}

	if len(gptResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return gptResp.Choices[0].Text, nil
}

// Function to format the valid actions for the GPT prompt.
func formatValidActions(validActions map[string]string) string {
	var sb strings.Builder
	sb.WriteString("Valid actions:\n")
	for action, usage := range validActions {
		sb.WriteString(fmt.Sprintf("%s: %s\n", action, usage))
	}
	return sb.String()
}

// Function to parse actions from GPT response.
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
