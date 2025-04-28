package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from Backend 1 (Host: %s)", r.Host)
	})
	fmt.Println("Backend 1 server running on :8081")
	http.ListenAndServe(":8081", nil)
}
