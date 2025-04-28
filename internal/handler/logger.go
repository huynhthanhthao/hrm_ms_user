package handler

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

var logger = log.Default()

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        t := time.Now()
        c.Next()
        latency := time.Since(t)
        logger.Print(latency)
        status := c.Writer.Status()
        logger.Println(status)
    }
}