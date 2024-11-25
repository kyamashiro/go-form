package session

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	MaxLifetime = 60 * 60 * 24 // 1日
	SId         = "s_id"
	Dir         = "./tmp/sessions" // セッションデータ保存ディレクトリ
)

type Manager struct {
	lock sync.Mutex
}

type Session struct {
	Values    map[string]interface{}
	Id        string
	ExpiresAt time.Time
	lock      sync.Mutex
}

// 新しいセッションマネージャを生成
func NewManager() (*Manager, error) {
	if err := os.MkdirAll(Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}
	return &Manager{}, nil
}

// セッションIDの生成
func generateId() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ファイルパスを取得
func filePath(sid string) string {
	return filepath.Join(Dir, sid)
}

// セッションの開始
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (*Session, error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	fmt.Println("SessionStart")
	// クッキーからセッションIDを取得
	cookie, err := r.Cookie(SId)
	if err != nil {
		println("cookie is nil")
		// クッキーが存在しない場合は新規セッションを生成
		sid, _ := generateId()
		session := &Session{
			Values:    make(map[string]interface{}),
			Id:        sid,
			ExpiresAt: time.Now().Add(time.Duration(MaxLifetime) * time.Second),
		}
		err := session.Save()
		if err != nil {
			return nil, err
		}

		http.SetCookie(w, &http.Cookie{
			Name:     SId,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   MaxLifetime,
			SameSite: http.SameSiteStrictMode,
		})

		return session, nil
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	session := load(sid)
	fmt.Println(session)
	if session == nil || time.Now().After(session.ExpiresAt) {
		println("session is nil")
		// セッションが無効の場合、新しいセッションを生成
		sid, _ = generateId()
		session = &Session{
			Values:    make(map[string]interface{}),
			Id:        sid,
			ExpiresAt: time.Now().Add(time.Duration(MaxLifetime) * time.Second),
		}
		err := session.Save()
		if err != nil {
			return nil, err
		}
		http.SetCookie(w, &http.Cookie{
			Name:     SId,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   MaxLifetime,
			SameSite: http.SameSiteStrictMode,
		})
	}
	return session, nil
}

// セッションをロード
func load(sid string) *Session {
	filePath := filePath(sid)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	session := &Session{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil
	}

	return session
}

// セッションの破棄
func (manager *Manager) Destroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SId)
	if err != nil || cookie.Value == "" {
		return
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	sid, _ := url.QueryUnescape(cookie.Value)
	filePath := filePath(sid)
	os.Remove(filePath) // セッションファイルを削除

	http.SetCookie(w, &http.Cookie{
		Name:     SId,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// セッションの保存
func (session *Session) Save() error {
	data, err := json.Marshal(session)
	fmt.Println("Saving session:", string(data))
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(Dir, session.Id), data, 0644)
	if err != nil {
		return err
	}
	return nil
}
