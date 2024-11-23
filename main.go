package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"go-form/controllers/signup"
	"html/template"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", signup.Index)
	r.Post("/sign-up", signup.Create)
	CSRF := csrf.Protect(
		[]byte("32-byte-long-auth-key"),
		csrf.CookieName("csrfToken"),
		csrf.RequestHeader("csrfToken"),
		csrf.FieldName("csrfToken"),
		csrf.Secure(false),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("CSRF攻撃の疑いのあるリクエストが発行されました")
			fmt.Fprintf(w, "(%s)", csrf.FailureReason(r).Error())
		})))
	err := http.ListenAndServe(":80", CSRF(r))
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
