package overlay

import (
	"fmt"
	"log"
	"net/http"
)

func MusicOverlay() {
	mux := http.NewServeMux()

	// Serve the HTML file directly
	mux.HandleFunc("/music", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := content.ReadFile("static/music.html")
		if err != nil {
			http.Error(w, "Error reading embedded HTML file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlContent)
	})

	// Serve the full-screen now playing overlay
	mux.HandleFunc("/nowplaying", func(w http.ResponseWriter, r *http.Request) {
		htmlContent, err := content.ReadFile("static/nowplaying.html")
		if err != nil {
			http.Error(w, "Error reading embedded HTML file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlContent)
	})

	// Serve static JS/CSS
	mux.Handle("/static/", http.FileServer(http.FS(content)))

	// WebSocket Upgrader
	mux.HandleFunc("/ws", handleMusicWebSocket)

	log.Printf("Starting MusicOverlay on port 38080\n")
	fmt.Printf("MusicOverlay starting on http://localhost:38080/music\n")

	err := http.ListenAndServe(":38080", mux)
	if err != nil {
		fmt.Printf("Error starting MusicOverlay server: %v\n", err)
	}
}
