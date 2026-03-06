package controllers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/calebchiang/thirdparty_server/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	openai "github.com/sashabaranov/go-openai"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func PracticeSocket(c *gin.Context) {

	// Upgrade HTTP → WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Get session_id from query
	sessionIDParam := c.Query("session_id")

	sessionID, err := strconv.Atoi(sessionIDParam)
	if err != nil {
		conn.WriteJSON(gin.H{
			"type":  "error",
			"error": "invalid session_id",
		})
		return
	}

	// Load practice session
	var session models.PracticeSession

	err = database.DB.First(&session, sessionID).Error
	if err != nil {
		conn.WriteJSON(gin.H{
			"type":  "error",
			"error": "session not found",
		})
		return
	}

	// Keep conversation in memory for the lifetime of this websocket
	var conversation []openai.ChatCompletionMessage

	// Generate first AI message
	firstMessage, err := services.GenerateFirstMessage(
		session.Scenario,
		session.Persona,
	)
	if err != nil {
		conn.WriteJSON(gin.H{
			"type":  "error",
			"error": "failed to generate AI message",
		})
		return
	}

	// Save first assistant message to DB
	message := models.PracticeMessage{
		SessionID: session.ID,
		Role:      "assistant",
		Content:   firstMessage,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		conn.WriteJSON(gin.H{
			"type":  "error",
			"error": "failed to save message",
		})
		return
	}

	// Add first assistant message to in-memory conversation
	conversation = append(conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: firstMessage,
	})

	// Generate speech using TTS
	audioBytes, err := services.GenerateSpeech(firstMessage, session.Persona)

	audioBase64 := ""

	if err != nil {
		fmt.Println("TTS generation failed:", err)
	} else {
		fmt.Println("Generated speech bytes:", len(audioBytes))

		// Convert audio bytes to Base64
		audioBase64 = base64.StdEncoding.EncodeToString(audioBytes)

		fmt.Println("Sending audio through websocket")
	}

	// Send AI message through websocket
	conn.WriteJSON(gin.H{
		"type":    "assistant_message",
		"content": firstMessage,
		"audio":   audioBase64,
	})

	for {
		var msg map[string]interface{}

		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("WebSocket closed:", err)
			break
		}

		msgType, _ := msg["type"].(string)

		if msgType == "user_message" {
			userText, _ := msg["content"].(string)

			if userText == "" {
				continue
			}

			// Save user message to DB
			if err := database.DB.Create(&models.PracticeMessage{
				SessionID: session.ID,
				Role:      "user",
				Content:   userText,
			}).Error; err != nil {
				fmt.Println("Failed to save user message:", err)
				continue
			}

			// Add user message to in-memory conversation
			conversation = append(conversation, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: userText,
			})

			// Optional: limit prompt size so the conversation does not grow forever
			if len(conversation) > 12 {
				conversation = conversation[len(conversation)-12:]
			}

			// Generate AI reply using in-memory conversation
			reply, err := services.GenerateReply(
				session.Scenario,
				session.Persona,
				conversation,
			)
			if err != nil {
				fmt.Println("Failed to generate reply:", err)
				continue
			}

			// Save assistant reply to DB
			if err := database.DB.Create(&models.PracticeMessage{
				SessionID: session.ID,
				Role:      "assistant",
				Content:   reply,
			}).Error; err != nil {
				fmt.Println("Failed to save assistant reply:", err)
				continue
			}

			// Add assistant reply to in-memory conversation
			conversation = append(conversation, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: reply,
			})

			// Optional: limit again after appending assistant reply
			if len(conversation) > 12 {
				conversation = conversation[len(conversation)-12:]
			}

			// Generate TTS
			audioBytes, err := services.GenerateSpeech(reply, session.Persona)

			audioBase64 := ""

			if err != nil {
				fmt.Println("TTS generation failed:", err)
			} else if len(audioBytes) > 0 {
				audioBase64 = base64.StdEncoding.EncodeToString(audioBytes)
			}

			// Send assistant reply back through websocket
			conn.WriteJSON(gin.H{
				"type":    "assistant_message",
				"content": reply,
				"audio":   audioBase64,
			})
		}
	}
}
