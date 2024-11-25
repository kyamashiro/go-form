package signup

import (
	"go-form/core/session"
	"html/template"
	"log/slog"
	"net/http"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	manager, err := session.NewManager()
	if err != nil {
		slog.Error("Error occurred: %v", err)
		http.Redirect(w, nil, "/500", http.StatusInternalServerError)
	}
	s, _ := manager.SessionStart(w, r)
	s.Values["csrfToken"] = ""
	s.Save()

	t, _ := template.ParseFiles("templates/sign_up.html")
	err := t.Execute(w, map[string]interface{}{
		"csrfToken": "",
	})
	if err != nil {
		slog.Error("Error occurred: %v", err)
		http.Redirect(w, nil, "/500", http.StatusInternalServerError)
	}
}

//func Create(w http.ResponseWriter, r *http.Request) {
//	userName := r.FormValue("userName")
//	password := r.FormValue("password")
//	fmt.Println(userName)
//	fmt.Println(password)
//	/**
//	* ここでユーザー登録処理を行う
//	1. ユーザー名とパスワードのバリデーション
//		1.1 バリデーションエラーの場合はエラーメッセージを表示
//			a. ユーザー名とパスワードが空でないか
//			b. パスワードが8文字以上か
//			c. ユーザー名が既に登録されていないか
//		1.2 バリデーションエラーがない場合は次の処理へ
//	2. ユーザー登録
//		2.1 パースワードをハッシュ化
//		2.2 ユーザー名とパスワードをDBに保存
//		2.3 認証処理を実行
//	3. ホーム画面にリダイレクト
//	**/
//	t, _ := template.ParseFiles("templates/sign_up.html")
//	err := t.Execute(w, nil)
//	exception.ErrorHandler.Report(w, err)
//}
