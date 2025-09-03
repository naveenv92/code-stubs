package main

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	r := gin.Default()
	r.GET("/ws", handleWs)

	slog.Info("listening to websocket connections at /ws...")
	r.Run(":8080")
}

func handleWs(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("error upgrading message: %v", err))
		return
	}
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				slog.Error(fmt.Sprintf("error while reading message: %v", err))
				continue
			}

			// close the connection gracefully
			slog.Info("gracefully shutdown of websocket connection")
			return
		}

		if err := conn.WriteMessage(messageType, p); err != nil {
			slog.Error(fmt.Sprintf("error while writing message: %v", err))
			continue
		}
	}
}
