package overlay

import (
	"embed"
	"fmt"
	"net/http"
)

//go:embed static
var content embed.FS

func handleIndex(w http.ResponseWriter, r *http.Request) {
	htmlContent, err := content.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Error reading embedded HTML file", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "text/html")

	// Write the HTML content to the response
	_, err = w.Write(htmlContent)
	if err != nil {
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func WebOverlay() {
	http.HandleFunc("/overlay", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, web!")
	})

	http.HandleFunc("/index", handleIndex)

	fmt.Printf("Starting /overlay on port 28080\n")
	err := http.ListenAndServe(":28080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
