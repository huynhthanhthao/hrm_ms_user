package helper

import (
	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
