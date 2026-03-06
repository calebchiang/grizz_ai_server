package services

import (
	"strings"

	"github.com/calebchiang/thirdparty_server/models"
)

func ReconstructTranscript(persona string, messages []models.PracticeMessage) string {

	var builder strings.Builder

	for _, m := range messages {

		if m.Role == "assistant" {

			builder.WriteString(persona)
			builder.WriteString(": ")
			builder.WriteString(m.Content)
			builder.WriteString("\n\n")

		} else {

			builder.WriteString("You: ")
			builder.WriteString(m.Content)
			builder.WriteString("\n\n")

		}
	}

	return builder.String()
}
