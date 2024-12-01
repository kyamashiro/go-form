package signup

import (
	"fmt"
	"go-form/core/database"
	"go-form/core/session"
	"go-form/repo"
	"html/template"
	"net/http"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		post(w, r)
	case http.MethodGet:
		get(w)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func get(w http.ResponseWriter) {
	t, _ := template.ParseFiles("template/sign_up.html")
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

/*
*
  - ここでユーザー登録処理を行う
    1. ユーザー名とパスワードのバリデーション
    1.1 バリデーションエラーの場合はエラーメッセージを表示
    a. ユーザー名とパスワードが空でないか
    b. パスワードが8文字以上か
    c. ユーザー名が既に登録されていないか
    1.2 バリデーションエラーがない場合は次の処理へ
    2. ユーザー登録
    2.1 パースワードをハッシュ化
    2.2 ユーザー名とパスワードをDBに保存
    2.3 認証処理を実行
    3. ホーム画面にリダイレクト

*
*/
func post(w http.ResponseWriter, r *http.Request) {
	errMsg, hasErr, err := validate(r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if hasErr {
		t, _ := template.ParseFiles("template/sign_up.html")
		fmt.Printf("%v", r.Form)
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
	db := database.DB()
	defer db.Close()
	userRepo := repo.NewUserRepository(db)
	// ユーザー登録
	user, err := userRepo.Create(r.FormValue("userName"), r.FormValue("password"))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// 認証処理を実行してホーム画面にリダイレクト
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
	s.Values["user"] = user
	err = s.Save()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
	return
}

func validate(r *http.Request) (map[string][]string, bool, error) {
	errMsg := make(map[string][]string)
	hasErr := false
	userName := r.FormValue("userName")
	password := r.FormValue("password")
	if userName == "" {
		errMsg["userName"] = append(errMsg["userName"], "ユーザー名は必須です")
		hasErr = true
	}

	if password == "" {
		errMsg["password"] = append(errMsg["password"], "パスワードは必須です")
		hasErr = true
	}
	if len(password) < 8 {
		errMsg["password"] = append(errMsg["password"], "パスワードは8文字以上で入力してください")
		hasErr = true
	}

	db := database.DB()
	defer db.Close()
	userRepo := repo.NewUserRepository(db)
	exists, err := userRepo.Exists(userName)
	if err != nil {
		return errMsg, hasErr, err
	}

	if exists {
		errMsg["userName"] = append(errMsg["userName"], "ユーザー名は既に登録されています")
		hasErr = true
	}

	return errMsg, hasErr, nil
}
