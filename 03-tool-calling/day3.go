package toolcalling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Tool definition — what you send to the API
type ToolParam struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  struct {
		Type       string               `json:"type"`
		Properties map[string]ToolParam `json:"properties"`
		Required   []string             `json:"required"`
	} `json:"parameters"`
}

type Tool struct {
	Type     string       `json:"type"` // always "function"
	Function ToolFunction `json:"function"`
}

var tools = []Tool{
	{
		Type: "function",
		Function: ToolFunction{
			Name: "get_current_time",
			Description: "Returns the current date and time. " +
				"Call this when the user asks about the current time or date.",
			Parameters: struct {
				Type       string               `json:"type"`
				Properties map[string]ToolParam `json:"properties"`
				Required   []string             `json:"required"`
			}{Type: "object", Properties: map[string]ToolParam{}, Required: []string{}},
		},
	},
	{
		Type: "function",
		Function: ToolFunction{
			Name: "check_availability",
			Description: "Check if a time slot is available for a meeting. " +
				"Call this when the user wants to schedule or book something.",
			Parameters: struct {
				Type       string               `json:"type"`
				Properties map[string]ToolParam `json:"properties"`
				Required   []string             `json:"required"`
			}{
				Type: "object",
				Properties: map[string]ToolParam{
					"date":          {Type: "string", Description: "Date in YYYY-MM-DD format"},
					"duration_mins": {Type: "integer", Description: "Meeting duration in minutes"},
				},
				Required: []string{"date", "duration_mins"},
			},
		},
	},
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

func callWithTools(msgs []Message) Message {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}

	body, _ := json.Marshal(map[string]interface{}{
		"model":    "llama3.2",
		"messages": msgs,
		"tools":    tools, // ← attach your tools here
		"stream":   false,
	})
	req, _ := http.NewRequest("POST", host+"/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var r struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(data, &r)
	return r.Choices[0].Message
}

// Stub implementations — replace with real logic later
func executeGetCurrentTime() string {
	return `{"time": "2026-03-30T10:30:00+05:30", "timezone": "IST"}`
}

func executeCheckAvailability(args string) string {
	var a struct {
		Date         string `json:"date"`
		DurationMins int    `json:"duration_mins"`
	}
	json.Unmarshal([]byte(args), &a)
	return fmt.Sprintf(
		`{"available": true, "slots": ["09:00", "14:00", "16:30"], "date": "%s", "duration_mins": %d}`,
		a.Date, a.DurationMins)
}

func ToolCaller() {
	msgs := []Message{
		{Role: "system", Content: "You are a scheduling assistant. Use tools when needed."},
		{Role: "user", Content: "Can you check if tomorrow has any 30-minute slots free?"},
	}

	reply := callWithTools(msgs)

	// Did the model want to call a tool?
	if len(reply.ToolCalls) > 0 {
		tc := reply.ToolCalls[0]
		fmt.Printf("Model wants to call: %s\nArgs: %s\n", tc.Function.Name, tc.Function.Arguments)

		// Execute the right tool
		var result string
		switch tc.Function.Name {
		case "get_current_time":
			result = executeGetCurrentTime()
		case "check_availability":
			result = executeCheckAvailability(tc.Function.Arguments)
		}

		// Append: assistant's tool call + tool result
		msgs = append(msgs, reply)
		msgs = append(msgs, Message{
			Role:       "tool",
			Content:    result,
			ToolCallID: tc.ID,
		})

		// Call again to get the final natural language answer
		final := callWithTools(msgs)
		fmt.Println("Final answer:", final.Content)
	} else {
		// Model answered directly without a tool
		fmt.Println("Direct answer:", reply.Content)
	}
}
