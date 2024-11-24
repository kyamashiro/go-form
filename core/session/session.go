package session

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const MaxLifetime = 60 * 60 * 24

const SId = "s_id"

type Manager struct {
	lock     sync.Mutex
	sessions map[string]Func
}

type Func interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Delete(key interface{}) error
	SessionID() string
}

type Session struct {
	values    map[interface{}]interface{}
	sid       string
	expiresAt time.Time
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]Func),
	}
}

// セッションIDの生成
func (manager *Manager) generateId() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) Func {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// クッキーからセッションIDを取得
	cookie, err := r.Cookie(SId)
	// クッキーが存在しない場合は新規セッションを生成
	if err != nil || cookie.Value == "" {
		sid := manager.generateId()
		session := &Session{values: make(map[interface{}]interface{}), sid: sid, expiresAt: time.Now().AddDate(0, 1, 0)}
		manager.sessions[sid] = session

		// セッションIDをクッキーにセット
		http.SetCookie(w, &http.Cookie{
			Name:     SId,
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   MaxLifetime,
			SameSite: http.SameSiteStrictMode,
		})
		return session
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	session, exists := manager.sessions[sid]
	// セッションIDが存在しないもしくは有効期限が切れている場合は新規セッションを生成
	if !exists || time.Now().After(session.(*Session).values["expiresAt"].(time.Time)) {
		sid = manager.generateId()
		session := &Session{values: make(map[interface{}]interface{}), sid: sid, expiresAt: time.Now().AddDate(0, 1, 0)}
		manager.sessions[sid] = session
		return session
	}
	return session
}

func (session *Session) Set(key, value interface{}) error {
	session.values[key] = value
	return nil
}

func (session *Session) Get(key interface{}) interface{} {
	return session.values[key]
}

func (session *Session) Delete(key interface{}) error {
	delete(session.values, key)
	return nil
}

func (session *Session) SessionID() string {
	return session.sid
}

func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SId)
	if err != nil || cookie.Value == "" {
		return
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()

	sid, _ := url.QueryUnescape(cookie.Value)
	delete(manager.sessions, sid)

	http.SetCookie(w, &http.Cookie{
		Name:     SId,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		// 有効期限:1ヶ月
		MaxAge: time.Now().AddDate(0, 1, 0).Second(),
	})
}

func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	for sid, session := range manager.sessions {
		// セッションの期限切れチェックを行い、必要に応じて削除
		// ここでは単純なタイムアウトを利用する例として実装
		if time.Now().Unix()-session.(*Session).values["last_access"].(int64) > MaxLifetime {
			delete(manager.sessions, sid)
		}
	}
}
