package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

type RevenueCatWebhookPayload struct {
	Event struct {
		Type      string `json:"type"`
		AppUserID string `json:"app_user_id"`
	} `json:"event"`
}

func RevenueCatWebhook(c *gin.Context) {

	var payload RevenueCatWebhookPayload

	// Parse webhook payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		fmt.Println("Failed to parse RevenueCat webhook:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	eventType := payload.Event.Type
	appUserID := payload.Event.AppUserID

	fmt.Println("====================================")
	fmt.Println("🔥 REVENUECAT WEBHOOK RECEIVED")
	fmt.Println("Event:", eventType)
	fmt.Println("User:", appUserID)
	fmt.Println("====================================")

	// Convert string user ID → uint
	userIDInt, err := strconv.Atoi(appUserID)
	if err != nil {
		fmt.Println("❌ Invalid app_user_id:", appUserID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	userID := uint(userIDInt)

	var user models.User

	if err := database.DB.First(&user, userID).Error; err != nil {
		fmt.Println("❌ User not found:", userID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	//--------------------------------
	// PURCHASE / RENEWAL EVENTS
	//--------------------------------

	if eventType == "INITIAL_PURCHASE" || eventType == "RENEWAL" {

		user.IsPremium = true
		user.Credits = 20

		if err := database.DB.Save(&user).Error; err != nil {
			fmt.Println("❌ Failed updating user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		fmt.Println("✅ User upgraded to premium:", userID)
	}

	//--------------------------------
	// EXPIRATION / CANCELLATION
	//--------------------------------

	if eventType == "EXPIRATION" || eventType == "CANCELLATION" {

		user.IsPremium = false
		user.Credits = 0
		user.SeenOnboarding = false

		if err := database.DB.Save(&user).Error; err != nil {
			fmt.Println("❌ Failed updating expiration:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		fmt.Println("⚠️ Premium expired for user:", userID)
	}

	//--------------------------------
	// Always return 200 to RevenueCat
	//--------------------------------

	c.JSON(http.StatusOK, gin.H{
		"status": "processed",
	})
}
