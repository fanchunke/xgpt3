package conversation

import (
	"context"
	"time"
)

type Session struct {
	// ID of the session.
	ID int `json:"id,omitempty"`
	// 用户Id
	UserID string `json:"user_id,omitempty"`
	// 会话是否开启
	Status bool `json:"status,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// DeletedAt holds the value of the "deleted_at" field.
	DeletedAt int `json:"deleted_at,omitempty"`
}

type Message struct {
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// 会话Id
	SessionID int `json:"session_id,omitempty"`
	// 消息发送者Id
	FromUserID string `json:"from_user_id,omitempty"`
	// 消息接收者Id
	ToUserID string `json:"to_user_id,omitempty"`
	// 消息内容
	Content string `json:"content,omitempty"`
	// SpouseID holds the value of the "spouse_id" field.
	SpouseID int `json:"spouse_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type Handler interface {
	// 创建会话
	CreateSession(ctx context.Context, userId string) (*Session, error)
	// 关闭会话
	CloseSession(ctx context.Context, userId string) error
	// 获取最近一次开启的会话
	GetLatestActiveSession(ctx context.Context, userId string) (*Session, error)
	// 创建消息
	CreateMessage(ctx context.Context, session *Session, fromUserId, toUserId string, content string) (*Message, error)
	// 创建配对消息
	CreateSpouseMessage(ctx context.Context, session *Session, fromUserId, toUserId, content string, spouse *Message) (*Message, error)
	// 获取会话内最近的消息列表
	ListLatestMessagesWithSpouse(ctx context.Context, session *Session, userId string, turns int) ([]*Message, error)
}
