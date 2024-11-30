package signin

import (
	"fmt"
	"go-form/core/database"
	"go-form/core/session"
	"go-form/repo"
	"html/template"
	"net/http"
)

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		errMsg, hasErr, err := validate(r)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if hasErr {
			t, _ := template.ParseFiles("template/sign_in.html")
			err := t.Execute(w, map[string]interface{}{
				"errMsg":   errMsg,
				"userName": r.FormValue("userName"),
				"password": r.FormValue("password"),
			})
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}

		// ログイン成功時の処理
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

		db := database.DB()
		defer db.Close()
		userRepo := repo.NewUserRepository(db)
		user := userRepo.FindByName(r.FormValue("userName"))
		fmt.Printf("check")
		s.Values["user"] = user
		err = s.Save()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t, _ := template.ParseFiles("template/sign_in.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func validate(r *http.Request) (map[string][]string, bool, error) {
	errMsg := make(map[string][]string)
	hasErr := false
	if r.FormValue("userName") == "" {
		errMsg["userName"] = append(errMsg["userName"], "ユーザー名を入力してください")
		hasErr = true
	}
	if r.FormValue("password") == "" {
		errMsg["password"] = append(errMsg["password"], "パスワードを入力してください")
		hasErr = true
	}

	if hasErr {
		return errMsg, hasErr, nil
	}

	db := database.DB()
	defer db.Close()
	userRepo := repo.NewUserRepository(db)
	auth := userRepo.Auth(r.FormValue("userName"), r.FormValue("password"))
	if auth == false {
		errMsg["password"] = append(errMsg["password"], "ログインに失敗しました")
		hasErr = true
	}
	return errMsg, false, nil
}
