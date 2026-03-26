package controllers

import (
	"net/http"

	"github.com/calebchiang/thirdparty_server/services"
	"github.com/gin-gonic/gin"
)

func UploadVideo(c *gin.Context) {

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Video file required",
		})
		return
	}

	videoURL, err := services.UploadVideoToR2(file)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload video",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video_url": videoURL,
	})
}
