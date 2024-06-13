package main

import (
	"fmt"
	"strings"
)

// Action represents a user action with its name and arguments.
type Action struct {
	Name string
	Args []string
}

// cleanUserQuery cleans the user's query by removing comments and handling special cases.
func cleanUserQuery(input string) string {
	var result strings.Builder
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		// Handles inline comments
		if idx := strings.Index(line, "#"); idx != -1 {
			// Exclude comments that aren't within a string literal
			if !strings.Contains(line[:idx], "'") && !strings.Contains(line[:idx], "\"") {
				line = line[:idx]
			}
		}
		if strings.TrimSpace(line) != "" {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(line)
		}
	}
	return result.String()
}

// formatValidActions formats the valid actions into a string, ensuring sorted order.
func formatValidActions(actions []Action) string {
	var sb strings.Builder
	sb.WriteString("Valid actions:")
	for _, a := range actions {
		sb.WriteString("\n" + a.Name + ": " + strings.Join(a.Args, " "))
	}
	return sb.String()
}

// parseActions parses the list of actions from a response string using a map of valid actions.
func parseActions(response string, valid []Action) ([]Action, error) {
	var actions []Action
	lines := strings.Split(response, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		actionName := fields[0]
		ok := false
		for _, v := range valid {
			if v.Name != actionName {
				continue
			}
			ok = true
			// Creating Args by considering handling for special cases like quotes
			args := parseArguments(fields[1:])
			actions = append(actions, Action{Name: actionName, Args: args})
		}
		if !ok {
			return nil, fmt.Errorf("invalid action: %s", actionName)
		}
	}

	return actions, nil
}

// parseArguments manages arguments, correctly grouping those enclosed in quotes.
func parseArguments(fields []string) []string {
	var args []string
	var currentArg strings.Builder
	inQuotes := false

	for _, field := range fields {
		startQuote := strings.HasPrefix(field, "\"") || strings.HasPrefix(field, "'")
		endQuote := strings.HasSuffix(field, "\"") || strings.HasSuffix(field, "'")

		if startQuote && endQuote && field != "\"\"" && field != "''" {
			args = append(args, field[1:len(field)-1])
			continue
		}

		if startQuote {
			inQuotes = true
			currentArg.WriteString(field[1:])
			continue
		}

		if endQuote {
			inQuotes = false
			currentArg.WriteString(" " + field[:len(field)-1])
			args = append(args, currentArg.String())
			currentArg.Reset()
			continue
		}

		if inQuotes {
			if currentArg.Len() > 0 {
				currentArg.WriteString(" ")
			}
			currentArg.WriteString(field)
		} else {
			args = append(args, field)
		}
	}

	// Handling case when last argument is still inside quotes
	if currentArg.Len() > 0 && inQuotes {
		args = append(args, currentArg.String())
	}

	return args
}
