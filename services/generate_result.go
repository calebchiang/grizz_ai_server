package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type ConversationScores struct {
	Clarity          int `json:"clarity"`
	Engagement       int `json:"engagement"`
	Confidence       int `json:"confidence"`
	ConversationFlow int `json:"conversation_flow"`
	SocialAwareness  int `json:"social_awareness"`
}

type ConversationResult struct {
	Scores            ConversationScores
	ConversationScore int
}

func buildScoringPrompt(transcript string) string {

	return fmt.Sprintf(`
You are evaluating a conversation between a USER and another person.

IMPORTANT:
In the transcript, lines beginning with "You:" represent the USER being evaluated.
Lines beginning with another name represent the other conversation partner.

Your job is to evaluate the USER's conversational ability.

Score the USER from 0 to 10 in the following categories:

clarity
engagement
confidence
conversation_flow
social_awareness

Scoring definitions:

Clarity
How clearly the user communicates their thoughts and ideas.

Engagement
How well the user keeps the conversation active and interesting by responding and asking questions.

Confidence
How confident and self-assured the user sounds when speaking.

Conversation Flow
How naturally and smoothly the conversation progresses with the user.

Social Awareness
How well the user responds appropriately to the context, tone, and social cues of the conversation.

Return ONLY valid JSON in this exact format:

{
 "clarity": number,
 "engagement": number,
 "confidence": number,
 "conversation_flow": number,
 "social_awareness": number
}

Do NOT include any explanation or text outside the JSON.

Transcript:
%s
`, transcript)
}

func GenerateConversationResult(transcript string) (*ConversationResult, error) {

	client, err := getClient()
	if err != nil {
		return nil, err
	}

	prompt := buildScoringPrompt(transcript)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT4oMini,
			Temperature: 0,
			MaxTokens:   120,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a strict evaluator of conversational ability.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content

	content = strings.TrimSpace(content)

	var scores ConversationScores

	err = json.Unmarshal([]byte(content), &scores)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI scoring response: %v\nResponse: %s", err, content)
	}

	// Calculate overall conversation score (0-100)
	total := scores.Clarity +
		scores.Engagement +
		scores.Confidence +
		scores.ConversationFlow +
		scores.SocialAwareness

	conversationScore := int(float64(total) / 50.0 * 100.0)

	result := ConversationResult{
		Scores:            scores,
		ConversationScore: conversationScore,
	}

	return &result, nil
}
