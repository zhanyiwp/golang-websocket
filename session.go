package main

import (
	"github.com/satori/go.uuid"
)

// Session struct
type Session struct {
	sID      string
	uID      string
	conn     *Conn
	settings map[string]interface{}
}

// NewSession create a new session
func NewSession(conn *Conn) (*Session, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	session := &Session{
		sID:      id.String(),
		uID:      "",
		conn:     conn,
		settings: make(map[string]interface{}),
	}
	return session, nil
}

// GetSessionID get session ID
func (s *Session) GetSessionID() string {
	return s.sID
}

// BindUserID bind a user ID to session
func (s *Session) BindUserID(uid string) {
	s.uID = uid
}

// GetUserID get user ID
func (s *Session) GetUserID() string {
	return s.uID
}

// GetConn get Conn pointer
func (s *Session) GetConn() *Conn {
	return s.conn
}

// SetConn set a Conn to session
func (s *Session) SetConn(conn *Conn) {
	s.conn = conn
}

// GetSetting get setting
func (s *Session) GetSetting(key string) interface{} {
	if v, ok := s.settings[key]; ok {
		return v
	}
	return nil
}

// SetSetting set setting
func (s *Session) SetSetting(key string, value interface{}) {
	s.settings[key] = value
}
