package controllers

import (
	"net/http"
	"strconv"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/calebchiang/thirdparty_server/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

	// Save message to DB
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

	// Send message through socket
	conn.WriteJSON(gin.H{
		"type":    "assistant_message",
		"content": firstMessage,
	})

	// Keep socket alive (no conversation logic yet)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
