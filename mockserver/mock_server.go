package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Args[1] // accepts port from CLI args

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Hello from backend %s\n", port)
	})

	fmt.Println("Mock server running on port ", port)
	http.ListenAndServe(":"+port, nil)
}
