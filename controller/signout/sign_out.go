package signout

import (
	"go-form/core/session"
	"net/http"
)

func SignOut(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		post(w, r)
	default:
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	manager, err := session.NewManager()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = manager.Destroy(w, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/home", http.StatusSeeOther)
	return
}
