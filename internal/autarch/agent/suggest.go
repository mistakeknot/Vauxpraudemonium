package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// QuestionSuggestion represents a suggested answer for an interview question
type QuestionSuggestion struct {
	QuestionID string
	Suggestion string
	Confidence float64 // 0-1, how confident the AI is in this suggestion
}

// SuggestAnswers generates suggested answers for interview questions based on the project description
func SuggestAnswers(ctx context.Context, agent *Agent, description string, questions []string) (map[string]string, error) {
	prompt := buildSuggestionPrompt(description, questions)

	// Set a reasonable timeout
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	resp, err := agent.Generate(ctx, GenerateRequest{
		Prompt: prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("agent generation failed: %w", err)
	}

	// Parse the response
	suggestions, err := parseSuggestionResponse(resp.Content, questions)
	if err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %w", err)
	}

	return suggestions, nil
}

func buildSuggestionPrompt(description string, questions []string) string {
	var sb strings.Builder

	sb.WriteString(`You are helping a user define their software project. Based on their brief description, suggest answers to these interview questions.

PROJECT DESCRIPTION:
`)
	sb.WriteString(description)
	sb.WriteString(`

For each question, provide a thoughtful suggested answer that the user can review and edit.
Be specific but concise. If you're uncertain about something, make a reasonable assumption and note it.

QUESTIONS TO ANSWER:
`)

	for i, q := range questions {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, q))
	}

	sb.WriteString(`
Output ONLY valid JSON in this exact format (no markdown, no explanation):
{
  "suggestions": {
    "vision": "Suggested answer for the vision question...",
    "users": "Suggested answer for the users question...",
    "problem": "Suggested answer for the problem question...",
    "platform": "Web|CLI|Desktop|Mobile|API/Backend",
    "language": "Go|TypeScript|Python|Rust|Other",
    "requirements": "Requirement 1\nRequirement 2\nRequirement 3"
  }
}

For "platform" and "language", choose the most appropriate option from the list.
For "requirements", list 3-7 key requirements, one per line.

Generate the JSON now:`)

	return sb.String()
}

func parseSuggestionResponse(content string, questions []string) (map[string]string, error) {
	// Clean up the response
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Try to find JSON in the response
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		content = content[start : end+1]
	}

	var response struct {
		Suggestions map[string]string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(content), &response); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w\nContent: %s", err, content[:min(500, len(content))])
	}

	if len(response.Suggestions) == 0 {
		return nil, fmt.Errorf("no suggestions generated")
	}

	return response.Suggestions, nil
}

// SuggestSingleAnswer generates a suggestion for a single question given context
func SuggestSingleAnswer(ctx context.Context, agent *Agent, description string, previousAnswers map[string]string, questionID, questionPrompt string) (string, error) {
	prompt := buildSingleSuggestionPrompt(description, previousAnswers, questionID, questionPrompt)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := agent.Generate(ctx, GenerateRequest{
		Prompt: prompt,
	})
	if err != nil {
		return "", fmt.Errorf("agent generation failed: %w", err)
	}

	// Clean up response - just return the text, no JSON parsing needed
	suggestion := strings.TrimSpace(resp.Content)
	// Remove any markdown or quotes
	suggestion = strings.Trim(suggestion, "\"'`")

	return suggestion, nil
}

func buildSingleSuggestionPrompt(description string, previousAnswers map[string]string, questionID, questionPrompt string) string {
	var sb strings.Builder

	sb.WriteString("You are helping define a software project. Suggest an answer for this question.\n\n")
	sb.WriteString("PROJECT DESCRIPTION:\n")
	sb.WriteString(description)
	sb.WriteString("\n\n")

	if len(previousAnswers) > 0 {
		sb.WriteString("ALREADY ANSWERED:\n")
		for k, v := range previousAnswers {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("QUESTION TO ANSWER:\n")
	sb.WriteString(questionPrompt)
	sb.WriteString("\n\n")

	sb.WriteString("Provide ONLY the suggested answer, nothing else. Be concise and specific.")

	return sb.String()
}
