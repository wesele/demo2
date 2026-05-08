// Package ai handles communication with the OpenAI-compatible chat completion API.
package ai

import "fmt"

const systemPromptTemplate = `You are a shell command expert. The user is on %s using %s.
Your task: translate the user's natural-language request into a single, correct shell command.
Rules:
- Output ONLY the raw shell command. No explanations, no markdown, no code fences.
- Do not wrap the command in backticks or quotes.
- If you cannot produce a safe and valid command, output: ERROR: <reason>`

// buildSystemPrompt returns the system prompt with OS and shell substituted.
func buildSystemPrompt(goos, shell string) string {
	return fmt.Sprintf(systemPromptTemplate, goos, shell)
}
