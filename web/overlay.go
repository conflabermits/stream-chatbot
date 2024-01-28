package overlay

import (
	"fmt"
	"net/http"
)

func WebOverlay() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, web!")
	})
	fmt.Printf("Starting server on port 8080\n")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
