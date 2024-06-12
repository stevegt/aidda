package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

// Mocking the GPT API
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Helper function to create a mock response
func MockResponse(body string, statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

// Test the queryGPT function
func TestQueryGPT(t *testing.T) {
	mockClient := new(MockHTTPClient)
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

	apiResponse := `{
		"choices": [
			{
				"text": "runTests packagePath"
			}
		]
	}`

	mockClient.On("Do", mock.Anything).Return(MockResponse(apiResponse, 200), nil)

	// Inject the mock client into the HTTP client used in queryGPT
	oldClient := http.DefaultClient
	http.DefaultClient = mockClient
	defer func() { http.DefaultClient = oldClient }()

	userInstruction := "Please run tests and show the results."
	response, err := queryGPT(userInstruction, validActions)
	assert.NoError(t, err)
	assert.Equal(t, "runTests packagePath", response)
}

// Test the cleanUserQuery function
func TestCleanUserQuery(t *testing.T) {
	input := `# GPT Message:
# Here is a message from GPT. Please provide your query based on this information.
runTests packagePath`

	expected := "runTests packagePath"
	result := cleanUserQuery(input)
	assert.Equal(t, expected, result)
}

// Mock user input for launchEditor
func MockLaunchEditor(template string) (string, error) {
	return "runTests packagePath", nil
}

// Test the main function workflow
func TestMainWorkflow(t *testing.T) {
	// Mock the editor launch
	oldLaunchEditor := launchEditor
	launchEditor = MockLaunchEditor
	defer func() { launchEditor = oldLaunchEditor }()

	// Mock the queryGPT call
	mockClient := new(MockHTTPClient)
	apiResponse := `{
		"choices": [
			{
				"text": "runTests packagePath"
			}
		]
	}`
	mockClient.On("Do", mock.Anything).Return(MockResponse(apiResponse, 200), nil)
	oldClient := http.DefaultClient
	http.DefaultClient = mockClient
	defer func() { http.DefaultClient = oldClient }()

	// Mock valid actions
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

	// Execute main workflow
	gptMessage := "Here is a message from GPT. Please provide your query based on this information."
	template := formatTemplate(gptMessage)
	userQuery, err := launchEditor(template)
	assert.NoError(t, err)

	// Clean up user input (remove comments)
	userQuery = cleanUserQuery(userQuery)

	// Forward the query to GPT-4o with valid actions
	gptResponse, err := queryGPT(userQuery, validActions)
	assert.NoError(t, err)
	assert.Equal(t, "runTests packagePath", gptResponse)

	// Parse actions from GPT-4o response
	actions := parseActions(gptResponse, validActions)
	expectedActions := []Action{{Name: "runTests", Args: []string{"packagePath"}}}
	assert.Equal(t, expectedActions, actions)

	// Mock action execution in container
	result, err := executeActionInContainer(Action{Name: "runTests", Args: []string{"packagePath"}})
	assert.NoError(t, err)
	assert.Contains(t, result, "ok") // Assuming tests pass

	// Execute actions
	results, err := executeActions(actions)
	assert.NoError(t, err)
	assert.Contains(t, results[0], "runTests: ok") // Assuming tests pass
}

