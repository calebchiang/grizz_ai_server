package services

import (
	"context"
	"io"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var personaVoices = map[string]openai.SpeechVoice{
	"oliver": openai.SpeechVoice("alloy"),
	"john":   openai.SpeechVoice("sage"),
	"sophia": openai.SpeechVoice("aria"),
	"trisha": openai.SpeechVoice("verse"),
}

func GenerateSpeech(text string, persona string) ([]byte, error) {

	client, err := getClient()
	if err != nil {
		return nil, err
	}

	voice := personaVoices[strings.ToLower(persona)]
	if voice == "" {
		voice = openai.SpeechVoice("alloy")
	}

	req := openai.CreateSpeechRequest{
		Model: "gpt-4o-mini-tts",
		Voice: voice,
		Input: text,
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

func SaveSpeechFile(audio []byte, filename string) error {

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(audio)
	return err
}
