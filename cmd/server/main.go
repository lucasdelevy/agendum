package main

import (
	"fmt"
	"log"
	"net/http"

	"agendum/internal/handlers"
)

func main() {
	http.HandleFunc("/users/create/", handlers.CreateUser)
	
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}