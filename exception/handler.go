package exception

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type errorHandlerInterface interface {
	Report(w http.ResponseWriter, err error)
}

var ErrorHandler errorHandlerInterface = errorHandlerFunc{}

type errorHandlerFunc struct{}

func (ErrorHandler errorHandlerFunc) Report(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("Error occurred: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		t, _ := template.ParseFiles("templates/500.html")
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}
}
