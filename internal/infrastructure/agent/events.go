package agent

import (
	"google.golang.org/adk/session"
)

// ExtractTextFromEvent extracts all text content from an event
func ExtractTextFromEvent(event *session.Event) string {
	if event == nil || event.Content == nil {
		return ""
	}

	var text string
	for _, part := range event.Content.Parts {
		if part.Text != "" {
			text += part.Text
		}
	}
	return text
}

// HasFunctionCall checks if the event contains a function call
func HasFunctionCall(event *session.Event) bool {
	if event == nil || event.Content == nil {
		return false
	}

	for _, part := range event.Content.Parts {
		if part.FunctionCall != nil {
			return true
		}
	}
	return false
}

// HasFunctionResponse checks if the event contains a function response
func HasFunctionResponse(event *session.Event) bool {
	if event == nil || event.Content == nil {
		return false
	}

	for _, part := range event.Content.Parts {
		if part.FunctionResponse != nil {
			return true
		}
	}
	return false
}

// GetFunctionCallName returns the name of the function being called, if any
func GetFunctionCallName(event *session.Event) string {
	if event == nil || event.Content == nil {
		return ""
	}

	for _, part := range event.Content.Parts {
		if part.FunctionCall != nil {
			return part.FunctionCall.Name
		}
	}
	return ""
}
