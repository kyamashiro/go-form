package main

import (
	"go-form/controller/csv"
	"go-form/controller/home"
	"go-form/controller/signin"
	"go-form/controller/signout"
	"go-form/controller/signup"
	"go-form/core/csrf"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204で空レスポンスを返す
	})
	mux.HandleFunc("/", home.Home)
	mux.HandleFunc("/sign-up", signup.SignUp)
	mux.HandleFunc("/sign-in", signin.SignIn)
	mux.HandleFunc("/sign-out", signout.SignOut)
	mux.HandleFunc("/csv", csv.Csv)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", csrf.Middleware(mux)); err != nil {
		log.Fatal(err)
	}
}
