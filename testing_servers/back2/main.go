package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from Backend 2 (Host: %s)", r.Host)
	})
	fmt.Println("Backend 2 server running on :8082")
	http.ListenAndServe(":8082", nil)
}
