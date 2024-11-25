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

type Func interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Delete(key string)
	Id() string
}

type Session struct {
	Values    map[string]interface{}
	Sid       string
	ExpiresAt time.Time
	lock      sync.Mutex
}

// 新しいセッションマネージャを生成
func NewManager() *Manager {
	err := os.MkdirAll(Dir, 0755)
	if err != nil {
		return nil
	}
	return &Manager{}
}

// セッションIDの生成
func (manager *Manager) generateId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// ファイルパスを取得
func filePath(sid string) string {
	return filepath.Join(Dir, sid)
}

// セッションの開始
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) Func {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	fmt.Println("SessionStart")
	// クッキーからセッションIDを取得
	cookie, err := r.Cookie(SId)
	if err != nil {
		println("cookie is nil")
		// クッキーが存在しない場合は新規セッションを生成
		sid := manager.generateId()
		session := &Session{
			Values:    make(map[string]interface{}),
			Sid:       sid,
			ExpiresAt: time.Now().Add(time.Duration(MaxLifetime) * time.Second),
		}
		session.Save()

		http.SetCookie(w, &http.Cookie{
			Name:     SId,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   MaxLifetime,
			SameSite: http.SameSiteStrictMode,
		})

		return session
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	session := load(sid)
	fmt.Println(session)
	if session == nil || time.Now().After(session.ExpiresAt) {
		println("session is nil")
		// セッションが無効の場合、新しいセッションを生成
		sid = manager.generateId()
		session = &Session{
			Values:    make(map[string]interface{}),
			Sid:       sid,
			ExpiresAt: time.Now().Add(time.Duration(MaxLifetime) * time.Second),
		}
		session.Save()
		http.SetCookie(w, &http.Cookie{
			Name:     SId,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   MaxLifetime,
			SameSite: http.SameSiteStrictMode,
		})
	}
	return session
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
func (session *Session) Save() {
	data, err := json.Marshal(session)
	fmt.Println(session)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(Dir, session.Sid), data, 0644)
	if err != nil {
		panic(err)
	}
}

// セッションデータのセット
func (session *Session) Set(key string, value interface{}) {
	session.lock.Lock()
	defer session.lock.Unlock()

	latest := load(session.Sid)
	// 最新のセッションデータを取得
	session.Values = latest.Values
	session.Values[key] = value
	session.Save()
}

// セッションデータの取得
func (session *Session) Get(key string) interface{} {
	session.lock.Lock()
	defer session.lock.Unlock()

	latest := load(session.Sid)

	return latest.Values[key]
}

// セッションデータの削除
func (session *Session) Delete(key string) {
	session.lock.Lock()
	defer session.lock.Unlock()

	delete(session.Values, key)
	session.Save()
}

// セッションIDを取得
func (session *Session) Id() string {
	return session.Sid
}
