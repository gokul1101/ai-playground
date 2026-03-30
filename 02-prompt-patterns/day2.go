package promptpatterns

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "io"
    "os"
)

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func chat(msgs []Message) string {
	host := os.Getenv("OLLAMA_HOST")
    if host == "" {
        fmt.Println("Error: OLLAMA_HOST env variable is not set")
        return ""
    }
	body, _ := json.Marshal(map[string]interface{}{
        "model":    "llama3.2",
        "messages": msgs,
        "stream":   false,
    })
    req, _ := http.NewRequest("POST", host+"/v1/chat/completions", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
    data, _ := io.ReadAll(resp.Body)
    var r struct {
        Choices []struct{ Message Message `json:"message"` } `json:"choices"`
    }
    json.Unmarshal(data, &r)
    return r.Choices[0].Message.Content
}

func PromptPatterns() {
    // Pattern 1: system prompt
    result := chat([]Message{
        {Role: "system", Content: "You are a strict Go code reviewer. "+
            "Reply with only: PASS or FAIL, then one sentence why."},
        {Role: "user", Content: "func add(a, b int) int { return a + b }"},
    })
    fmt.Println("System prompt result:", result)

    // Pattern 2: few-shot — classify Go error severity
    result = chat([]Message{
        {Role: "system", Content: "Classify Go errors as CRITICAL, WARNING, or INFO."},

        // Example 1
        {Role: "user",      Content: "panic: runtime error: index out of range"},
        {Role: "assistant", Content: "CRITICAL"},

        // Example 2
        {Role: "user",      Content: "context deadline exceeded"},
        {Role: "assistant", Content: "WARNING"},

        // Example 3
        {Role: "user",      Content: "request completed in 142ms"},
        {Role: "assistant", Content: "INFO"},

        // Real input
        {Role: "user", Content: "logging silently and api failed"},
    })
    fmt.Println("Few-shot result:", result) // expects: CRITICAL

    // Pattern 3a: simple CoT
    result = chat([]Message{
        {Role: "system", Content: "You are a Go performance expert."},
        {Role: "user", Content: "Which is faster in Go: a map lookup "+
            "or a linear scan of a 10-element slice? Think step by step."},
    })
    fmt.Println(result)

    // Pattern 3b: structured reasoning — extract just the answer
    result2 := chat([]Message{
        {Role: "system", Content: "Reason inside <thinking> tags. "+
            "Then give your final answer after </thinking>."},
        {Role: "user", Content: "Should I use a goroutine or a worker pool "+
            "for 10,000 concurrent HTTP requests?"},
    })
    fmt.Println(result2)

    // Technique 1: instruct in system prompt (works 90% of the time)
    result = chat([]Message{
        {Role: "system", Content: "You are a meeting parser. "+
            "Always respond with valid JSON only. No markdown, no explanation.\n"+
            "Schema: {\"duration_mins\": int, \"participants\": []string, \"topic\": string}"},
        {Role: "user", Content: "Schedule a 30-minute sync with Alice and Bob about Q3 planning"},
    })

    // Parse it into a Go struct
    var meeting struct {
        DurationMins int      `json:"duration_mins"`
        Participants []string `json:"participants"`
        Topic        string   `json:"topic"`
    }
    if err := json.Unmarshal([]byte(result), &meeting); err != nil {
        fmt.Println("parse error:", err, "raw:", result)
        return
    }
    fmt.Printf("Duration: %d mins, Participants: %v\n",
        meeting.DurationMins, meeting.Participants)
}
