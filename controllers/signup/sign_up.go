package signup

import (
	"fmt"
	"go-form/core/database"
	"html/template"
	"net/http"
)

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
func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		errMsg, hasErr := validate(r)

		if hasErr {
			fmt.Println(errMsg)
			t, _ := template.ParseFiles("templates/sign_up.html")
			err := t.Execute(w, map[string]interface{}{
				"errMsg": errMsg,
			})
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}

	}

	t, _ := template.ParseFiles("templates/sign_up.html")
	err := t.Execute(w, map[string]interface{}{
		"hoge": "",
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func validate(r *http.Request) (map[string][]string, bool) {
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

	row := db.QueryRow("SELECT * FROM users WHERE name = $1", userName)
	fmt.Println(row.Scan())

	return errMsg, hasErr
}
