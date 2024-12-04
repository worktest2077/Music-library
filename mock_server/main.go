package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"releaseDate": "16.07.2006",
			"text":        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?",
			"link":        "https://www.youtube.com/watch?v=OgvLej8Trtc",
		})
	})

	r.Run(":8082")
}
