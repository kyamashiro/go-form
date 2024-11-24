package session

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	SessionID string
	ExpiresAt time.Time
}

type SessionStore struct {
	sessions map[string]map[string]interface{} // セッションID -> 任意のデータ
	mu       sync.Mutex
}

// セッションに任意のデータを設定
func (store *SessionStore) Set(sessionID string, key string, value interface{}) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// セッションIDが存在しない場合、新しいマップを作成
	if _, exists := store.sessions[sessionID]; !exists {
		store.sessions[sessionID] = make(map[string]interface{})
	}

	// セッションにデータをセット
	store.sessions[sessionID][key] = value
}

// セッションから任意のデータを取得
func (store *SessionStore) Get(sessionID string, key string) (interface{}, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// セッションIDが存在し、該当のキーがあればデータを返す
	if session, exists := store.sessions[sessionID]; exists {
		if value, ok := session[key]; ok {
			return value, true
		}
	}
	return nil, false
}

// セッションを削除
func (store *SessionStore) Delete(sessionID string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.sessions, sessionID)
}

// セッションIDをCookieから取得する
func getSessionId(r *http.Request) string {
	cookie, err := r.Cookie("s_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (store *SessionStore) init(w http.ResponseWriter, r *http.Request) (string, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	sessionId := getSessionId(r)

	if session, exists := store.sessions[sessionId]; exists {
		if expiresAt, ok := session["expiresAt"]; ok {
			// セッションの有効期限が切れている場合は削除
			if expiresAt.(time.Time).Before(time.Now()) {
				delete(store.sessions, sessionId)
			}
		}
	}

	if cookie.Value == "" {
		sessionId, err := generateSessionId()
		if err != nil {
			panic(err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "s_id",
			Value: sessionId,
			Path:  "/",
		})
	}

}

func generateSessionId() (string, error) {
	b := make([]byte, 128)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Base64エンコードして文字列に変換
	sessionID := base64.URLEncoding.EncodeToString(b)
	return sessionID, nil
}
