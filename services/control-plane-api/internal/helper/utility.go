package helper

import "github.com/gin-gonic/gin"

func RespondError(c *gin.Context, status int, code string, msg string) {
	// Standard error envelope used across handlers.
	c.JSON(status, gin.H{
		"error_code": code,
		"message": msg,
	})
}
