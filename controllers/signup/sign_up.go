package signup

import (
	"csv/exception"
	"fmt"
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/sign_up.html")
	err := t.Execute(w, nil)
	exception.ErrorHandler.Report(w, err)
}

func Create(w http.ResponseWriter, r *http.Request) {
	userName := r.FormValue("userName")
	password := r.FormValue("password")
	fmt.Println(userName)
	fmt.Println(password)
	t, _ := template.ParseFiles("templates/sign_up.html")
	err := t.Execute(w, nil)
	exception.ErrorHandler.Report(w, err)
}
