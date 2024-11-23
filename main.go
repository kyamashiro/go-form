package main

import (
	"csv/controllers/signup"
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

	r.Get("/", signup.Index)
	//r.Post("/", SignUp)

	err := http.ListenAndServe(":80", r)
	catch(err)
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/sign_in.html")
	err := t.Execute(w, nil)
	catch(err)
}

func catch(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
