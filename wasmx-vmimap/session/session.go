package session

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewExpirationTime() time.Time {
	return time.Now().Add(24 * time.Hour)
}

// / GenerateToken creates a JWT for the given username.
func GenerateToken(jwtSecret []byte, username string) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(NewExpirationTime()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// contextWithSession adds session to the request context
func ContextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}

// getSessionFromContext retrieves session from context
func GetSessionFromContext(ctx context.Context) (*Session, error) {
	session, ok := ctx.Value(SessionContextKey).(*Session)
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func GetValidSession(token string) (Session, error) {
	sessions.mu.Lock()
	session, exists := sessions.sessions[token]
	sessions.mu.Unlock()
	if !exists || session.Expires.Before(time.Now()) {
		return Session{}, fmt.Errorf("unauthorized: Invalid or expired session")
	}
	return session, nil
}

func UpdateAppSettings(token string, cb func(appSetings interface{}) interface{}) {
	sessions.mu.Lock()
	currentSession := sessions.sessions[token]

	currentSession.AppSettings = cb(currentSession.AppSettings)
	sessions.sessions[token] = currentSession
	sessions.mu.Unlock()
}

func DeleteSession(token string) {
	sessions.mu.Lock()
	delete(sessions.sessions, token)
	sessions.mu.Unlock()
}

func ReplaceSession(jwtSecret []byte, session Session) (Session, error) {
	oldToken := session.JWTToken
	session, err := BuildSession(jwtSecret, session)
	if err != nil {
		return Session{}, err
	}
	sessions.mu.Lock()
	sessions.sessions[session.JWTToken] = session
	delete(sessions.sessions, oldToken)
	sessions.mu.Unlock()
	return session, nil
}

func CreateSession(jwtSecret []byte, session Session) (Session, error) {
	session, err := BuildSession(jwtSecret, session)
	if err != nil {
		return Session{}, err
	}
	sessions.mu.Lock()
	sessions.sessions[session.JWTToken] = session
	sessions.mu.Unlock()
	return session, nil
}

func BuildSession(jwtSecret []byte, session Session) (Session, error) {
	if session.Username == "" {
		return Session{}, fmt.Errorf("create session: username missing")
	}
	token, err := GenerateToken(jwtSecret, session.Username)
	if err != nil {
		return Session{}, err
	}
	session.JWTToken = token
	session.Expires = NewExpirationTime()
	return session, nil
}

// func CreateSession(session Session) (string, Session) {
// 	session.Expires = time.Now().Add(24 * time.Hour) // 24-hour session

// 	// Create session
// 	sessionID := uuid.New().String()
// 	sessions.mu.Lock()
// 	sessions.sessions[sessionID] = session
// 	sessions.mu.Unlock()
// 	return sessionID, session
// }
