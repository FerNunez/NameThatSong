package server

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	// Add guess track route
	router.Post("/guess-track", handler.GuessTrack)

	// Add skip song route
	router.Post("/skip-song", handler.SkipSong)

	// Add clear queue route
	router.Post("/clear-queue", handler.ClearQueue)

	// Serve static files
	// ... existing code ...
}
