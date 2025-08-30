package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.NoMethod(methodNotFound)

	r.GET("/", sayHello)
	r.Run(":8080")
}

func methodNotFound(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "method not allowed"})
}

func sayHello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello!"})
}
