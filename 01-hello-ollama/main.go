package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func main() {
	host := os.Getenv("OLLAMA_HOST")
	model := os.Getenv("OLLAMA_MODEL")
	if host == "" {
		host = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}

	req := ChatRequest{
		Model:  model,
		Stream: true,
		Messages: []Message{
			{Role: "system", Content: "You are a helpful Go tutor."},
			{Role: "user", Content: "Explain Go interfaces in one sentence."},
		},
	}

	body, _ := json.Marshal(req)

	// Only difference from a cloud API: this URL
	url := host + "/v1/chat/completions"
	httpReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	// No Authorization header needed — Ollama is local

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Println("error — is Ollama running?", err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		chunk := strings.TrimPrefix(line, "data: ")
		if chunk == "[DONE]" {
			break
		}

		var delta struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(chunk), &delta) == nil {
			fmt.Print(delta.Choices[0].Delta.Content)
		}
	}
	fmt.Println()
}
