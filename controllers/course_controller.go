package controllers

import (
	"net/http"

	"github.com/calebchiang/thirdparty_server/database"
	"github.com/calebchiang/thirdparty_server/models"
	"github.com/gin-gonic/gin"
)

type CourseResponse struct {
	ID          uint   `json:"ID"`
	Category    string `json:"Category"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	SortOrder   int    `json:"SortOrder"`
	IsPublished bool   `json:"IsPublished"`
}

func GetCourses(c *gin.Context) {

	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var courses []models.Course

	if err := database.DB.
		Where("is_published = ?", true).
		Order("category ASC, sort_order ASC").
		Find(&courses).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load courses"})
		return
	}

	var response []CourseResponse

	for _, course := range courses {
		response = append(response, CourseResponse{
			ID:          course.ID,
			Category:    course.Category,
			Title:       course.Title,
			Description: course.Description,
			SortOrder:   course.SortOrder,
			IsPublished: course.IsPublished,
		})
	}

	c.JSON(http.StatusOK, response)
}
