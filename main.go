package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	//r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	//	w.Write([]byte("hello world"))
	//	http.ServeFile(w, r, "templates/login.html")
	//})

	r.Get("/", Login)

	err := http.ListenAndServe(":80", r)
	catch(err)
}

func Login(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/login.html")
	err := t.Execute(w, nil)
	catch(err)
}

func catch(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
