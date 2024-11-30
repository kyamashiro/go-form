package home

import (
	"go-form/core/session"
	"html/template"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	default:
		get(w, r)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	manager, err := session.NewManager()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s, err := manager.SessionStart(w, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	user := s.Values["user"]
	t, _ := template.ParseFiles("template/home.html")
	err = t.Execute(w, map[string]interface{}{
		"user": user,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
