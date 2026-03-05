package services

import (
	"context"
	"io"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var personaVoices = map[string]openai.SpeechVoice{
	"oliver": openai.SpeechVoice("alloy"),
	"john":   openai.SpeechVoice("echo"),
	"sophia": openai.SpeechVoice("marin"),
	"trisha": openai.SpeechVoice("sage"),
}

var personaInstructions = map[string]string{
	"oliver": "Speak in an outgoing and energetic tone with natural conversational pacing. Sound enthusiastic, confident, and engaging.",
	"john":   "Speak in a professional and composed tone with clear, confident delivery and steady conversational pacing.",
	"sophia": "Speak in a friendly and empathetic tone with warmth and natural conversational pacing.",
	"trisha": "Speak in a thoughtful and introverted tone with a calm, reflective style while maintaining natural conversational pacing.",
}

func GenerateSpeech(text string, persona string) ([]byte, error) {

	client, err := getClient()
	if err != nil {
		return nil, err
	}

	personaKey := strings.ToLower(persona)

	voice := personaVoices[personaKey]
	if voice == "" {
		voice = openai.SpeechVoice("alloy")
	}

	instructions := personaInstructions[personaKey]
	if instructions == "" {
		instructions = "Speak naturally in a conversational tone."
	}

	req := openai.CreateSpeechRequest{
		Model:        "gpt-4o-mini-tts",
		Voice:        voice,
		Input:        text,
		Instructions: instructions,
	}

	resp, err := client.CreateSpeech(
		context.Background(),
		req,
	)

	if err != nil {
		return nil, err
	}

	defer resp.Close()

	audioBytes, err := io.ReadAll(resp)
	if err != nil {
		return nil, err
	}

	return audioBytes, nil
}
