package main

import (
	"fmt"
	"go-form/controllers/signup"
	"go-form/core/csrf"
	"go-form/core/database"
	"log"
	"net/http"
)

type User struct {
	id        string
	name      string
	password  string
	createdAt string
	updatedAt string
}

func main() {
	mux := http.NewServeMux()

	db := database.DB()
	defer db.Close()

	row := db.QueryRow("SELECT * FROM users WHERE name = $1", "abc")
	u := &User{}
	if err := row.Scan(&u.id, &u.name, &u.password, &u.createdAt, &u.updatedAt); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", u)

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent) // 204で空レスポンスを返す
	})
	mux.HandleFunc("/", signup.SignUp)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", csrf.Middleware(mux)); err != nil {
		log.Fatal(err)
	}
}
