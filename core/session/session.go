package session

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
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
	Set(key, value interface{})
	Get(key interface{}) interface{}
	Delete(key interface{})
	Id() string
}

type Session struct {
	values    map[interface{}]interface{}
	sid       string
	expiresAt time.Time
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
func (manager *Manager) filePath(sid string) string {
	return filepath.Join(Dir, sid)
}

// セッションの開始
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) Func {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// クッキーからセッションIDを取得
	cookie, err := r.Cookie(SId)
	if err != nil || cookie.Value == "" {
		// クッキーが存在しない場合は新規セッションを生成
		sid := manager.generateId()
		session := &Session{
			values:    make(map[interface{}]interface{}),
			sid:       sid,
			expiresAt: time.Now().Add(time.Duration(MaxLifetime) * time.Second),
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
	session := manager.load(sid)
	if session == nil || time.Now().After(session.expiresAt) {
		// セッションが無効の場合、新しいセッションを生成
		sid = manager.generateId()
		session = &Session{
			values:    make(map[interface{}]interface{}),
			sid:       sid,
			expiresAt: time.Now().Add(time.Duration(MaxLifetime) * time.Second),
		}
		session.Save()
	}
	return session
}

// セッションをロード
func (manager *Manager) load(sid string) *Session {
	filePath := manager.filePath(sid)
	file, err := os.Open(filePath)
	if err != nil {
		return nil // ファイルが存在しない場合はnilを返す
	}
	defer file.Close()

	session := &Session{}
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(session); err != nil {
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
	filePath := manager.filePath(sid)
	os.Remove(filePath) // セッションファイルを削除

	http.SetCookie(w, &http.Cookie{
		Name:     SId,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func open(sid string) *Session {
	filePath := filepath.Join(Dir, sid)
	file, err := os.Open(filePath)
	if err != nil {
		return nil // ファイルが存在しない場合はnilを返す
	}
	defer file.Close()

	session := &Session{}
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(session); err != nil {
		return nil
	}

	return session
}

// セッションの保存
func (session *Session) Save() {
	data, err := json.Marshal(session)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(filepath.Join(Dir, session.sid), data, 0644)
	if err != nil {
		panic(err)
	}
}

// セッションデータのセット
func (session *Session) Set(key, value interface{}) {
	session.lock.Lock()
	defer session.lock.Unlock()

	latest := open(session.sid)
	// 最新のセッションデータを取得
	session.values = latest.values
	session.values[key] = value
	session.Save()
}

// セッションデータの取得
func (session *Session) Get(key interface{}) interface{} {
	session.lock.Lock()
	defer session.lock.Unlock()

	latest := open(session.sid)

	return latest.values[key]
}

// セッションデータの削除
func (session *Session) Delete(key interface{}) {
	session.lock.Lock()
	defer session.lock.Unlock()

	delete(session.values, key)
	session.Save()
}

// セッションIDを取得
func (session *Session) Id() string {
	return session.sid
}
