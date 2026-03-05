package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var personaPrompts = map[string]string{
	"oliver": `You are Oliver — outgoing, energetic, and confident.
You enjoy meeting new people and keeping conversations lively.`,

	"john": `You are John — professional, calm, and composed.
You speak clearly and thoughtfully.`,

	"sophia": `You are Sophia — friendly, warm, and empathetic.
You make people feel comfortable and understood.`,

	"trisha": `You are Trisha — introverted, thoughtful, and observant.
You speak calmly and ask meaningful questions.`,
}

func getClient() (*openai.Client, error) {

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	client := openai.NewClient(apiKey)

	return client, nil
}

func buildSystemPrompt(scenario string, persona string) string {

	key := strings.ToLower(persona)

	personaPrompt, ok := personaPrompts[key]
	if !ok {
		personaPrompt = personaPrompts["sophia"]
	}

	return fmt.Sprintf(`You are roleplaying a realistic conversation partner.

SCENARIO:
%s

PERSONA:
%s

CONVERSATION RULES:
- Speak naturally like a real human.
- Keep responses short and conversational (1–2 sentences).
- Ask follow-up questions to keep the conversation going.
- Do NOT explain the scenario.
- Do NOT break character.
- Never say you are an AI.

Start the interaction naturally as if you just met the user.`, scenario, personaPrompt)
}

func GenerateFirstMessage(
	scenario string,
	persona string,
) (string, error) {

	client, err := getClient()
	if err != nil {
		return "", err
	}

	systemPrompt := buildSystemPrompt(scenario, persona)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT4oMini,
			Temperature: 0.8,
			MaxTokens:   120,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Start the conversation.",
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

	message := resp.Choices[0].Message.Content

	return message, nil
}

func GenerateReply(
	scenario string,
	persona string,
	conversation []openai.ChatCompletionMessage,
) (string, error) {

	client, err := getClient()
	if err != nil {
		return "", err
	}

	systemPrompt := buildSystemPrompt(scenario, persona)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	messages = append(messages, conversation...)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT4oMini,
			Temperature: 0.8,
			MaxTokens:   120,
			Messages:    messages,
		},
	)

	if err != nil {
		return "", err
	}

	reply := resp.Choices[0].Message.Content

	return reply, nil
}
