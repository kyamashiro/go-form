package main

import (
	"go-form/controllers/signup"
	"go-form/core/csrf"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204で空レスポンスを返す
	})
	mux.HandleFunc("/", signup.SignUp)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", csrf.Middleware(mux)); err != nil {
		log.Fatal(err)
	}
}
