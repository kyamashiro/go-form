package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"go-form/controllers/signup"
	"log"
	"net/http"
)

func GenerateCSRFToken() (string, error) {
	// Generate 32 random bytes
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the token in a URL-safe base64 format
	return base64.URLEncoding.EncodeToString(token), nil
}

func main() {
	// csrfTokenを生成
	//csrfToken, err := GenerateCSRFToken()
	//if err != nil {
	//	fmt.Println("Error generating CSRF token:", err)
	//	return
	//}
	// SessionにcsrfTokenを保存
	//session.Set("csrfToken", csrfToken)

	http.HandleFunc("/", signup.SignUp)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
