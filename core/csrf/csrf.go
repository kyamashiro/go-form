package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"go-form/core/session"
	"net/http"
)

const name = "csrfToken"

func generate() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(token), nil
}

func validate(r *http.Request, sessionToken string) error {
	// リクエストからトークンを取得
	requestToken := r.FormValue(name) // フォームやクエリから取得
	if requestToken == "" {
		cookie, err := r.Cookie(name)
		if err != nil {
			return fmt.Errorf("CSRF token missing")
		}
		requestToken = cookie.Value
	}

	// トークンを比較
	if sessionToken != requestToken {
		return fmt.Errorf("invalid CSRF token")
	}
	return nil
}

// CSRFMiddleware https://blog.jxck.io/entries/2024-04-26/csrf.html
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// CSRF トークンをセッションから取得、なければ生成
		csrfToken, ok := s.Values[name].(string)
		if !ok || csrfToken == "" {
			csrfToken, err = generate()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			s.Values[name] = csrfToken
			if err := s.Save(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
		// CSRF トークンをレスポンスでクッキーとして返す
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    csrfToken,
			Path:     "/",
			HttpOnly: false, // クライアント側で読み取り可能にする
			SameSite: http.SameSiteStrictMode,
		})

		// POST リクエストの場合はトークンを検証
		if r.Method == http.MethodPost {
			if err := validate(r, csrfToken); err != nil {
				http.Error(w, "Forbidden: Invalid CSRF Token", http.StatusForbidden)
				return
			}
		}

		// 次のハンドラーを呼び出す
		next.ServeHTTP(w, r)
	})
}
