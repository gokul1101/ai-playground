package main

import (
	"github.com/joho/godotenv"
	// promptpatterns "github.com/gokul/ai-playground/02-prompt-patterns"
	toolcalling "github.com/gokul/ai-playground/03-tool-calling"
)

func main() {
	godotenv.Load(".env")  // loads .env into os environment
	// promptpatterns.PromptPatterns()
	toolcalling.ToolCaller()
}