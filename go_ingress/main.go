package main

import (
	"fmt"
	"net/http"
)

func main() {
	// This where the main application logic would go
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
