package main

import (
	"github.com/joho/godotenv"
	promptpatterns "github.com/gokul/ai-playground/02-prompt-patterns"
)

func main() {
	godotenv.Load(".env")  // loads .env into os environment
	promptpatterns.PromptPatterns()
}