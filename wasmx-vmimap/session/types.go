package session

import (
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// Claims defines the JWT claims we’ll include.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type SessionContextKeyType string

const SessionContextKey SessionContextKeyType = "session"

// Session represents a user session
type Session struct {
	Username    string
	Password    string
	Expires     time.Time
	AppSettings interface{}
	Token       *oauth2.Token
	Provider    string
	JWTToken    string
}

// SessionPublic represents a user session that can be made public
type SessionPublic struct {
	Username    string
	Expires     time.Time
	AppSettings interface{}
}

func (s *SessionPublic) FromPrivate(p *Session) {
	s.Username = p.Username
	s.Expires = p.Expires
	s.AppSettings = p.AppSettings
}

// SessionStore manages in-memory sessions
type SessionStore struct {
	sessions map[string]Session
	mu       sync.Mutex
}

var sessions = SessionStore{
	sessions: make(map[string]Session),
}

// func (s SessionStore) GetSessions() map[string]Session {
// 	return s.sessions
// }
