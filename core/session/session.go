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
	mu sync.RWMutex
}

type Session struct {
	Values    map[string]interface{}
	Id        string
	ExpiresAt time.Time
	mu        sync.RWMutex
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
	cookie, err := r.Cookie(SId)
	var sid string
	if err == nil {
		sid, _ = url.QueryUnescape(cookie.Value)
		session, err := manager.load(sid)
		if err != nil {
			return nil, err
		}

		if session != nil && !session.isExpired() {
			return session, nil
		}
	}

	// 新規セッションを生成
	newSid, err := generateId()
	fmt.Println(newSid)
	if err != nil {
		return nil, err
	}

	session := &Session{
		Values:    make(map[string]interface{}),
		Id:        newSid,
		ExpiresAt: time.Now().Add(time.Duration(MaxLifetime) * time.Second),
	}

	if err := session.Save(); err != nil {
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SId,
		Value:    url.QueryEscape(newSid),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   MaxLifetime,
		SameSite: http.SameSiteStrictMode,
	})

	return session, nil
}

// セッションをロード
func (manager *Manager) load(sid string) (*Session, error) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	// ファイルから読み込み
	filePath := filePath(sid)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	session := &Session{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil, err
	}

	return session, nil
}

// セッションの破棄
func (manager *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(SId)
	if err != nil || cookie.Value == "" {
		return err
	}

	sid, _ := url.QueryUnescape(cookie.Value)

	manager.mu.Lock()
	defer manager.mu.Unlock()
	err = os.Remove(filePath(sid))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SId,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	return nil
}

// セッションの保存
func (session *Session) Save() error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	session.mu.Lock()
	defer session.mu.Unlock()
	err = os.WriteFile(filePath(session.Id), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write session to file: %w", err)
	}
	return nil
}

// セッションの有効期限をチェック
func (session *Session) isExpired() bool {
	return time.Now().After(session.ExpiresAt)
}
