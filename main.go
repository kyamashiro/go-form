package main

import (
	"go-form/controllers/signup"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", signup.SignUp)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
