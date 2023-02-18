package ent

import (
	"context"
	"fmt"
	"sort"

	"github.com/fanchunke/xgpt3/conversation"
	"github.com/fanchunke/xgpt3/conversation/ent/chatent"
	"github.com/fanchunke/xgpt3/conversation/ent/chatent/message"
	"github.com/fanchunke/xgpt3/conversation/ent/chatent/session"
)

type ConversationHandler struct {
	client *chatent.Client
}

func New(client *chatent.Client) *ConversationHandler {
	return &ConversationHandler{client: client}
}

func (c *ConversationHandler) CreateSession(ctx context.Context, userId string) (*conversation.Session, error) {
	result, err := c.client.Session.
		Create().
		SetUserID(userId).
		SetStatus(true).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("Create Session failed: %w", err)
	}
	return toConversationSession(result), nil
}

func (c *ConversationHandler) CloseSession(ctx context.Context, userId string) error {
	_, err := c.client.Session.
		Update().
		Where(session.UserIDEQ(userId), session.StatusEQ(true)).
		SetStatus(false).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("Close User %s Session failed: %w", userId, err)
	}
	return nil
}

func (c *ConversationHandler) GetLatestActiveSession(ctx context.Context, userId string) (*conversation.Session, error) {
	result, err := c.client.Session.
		Query().
		Where(session.UserIDEQ(userId), session.StatusEQ(true)).
		Order(chatent.Desc(session.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetLatestActiveSession failed: %w", err)
	}
	return toConversationSession(result), nil
}

func (c *ConversationHandler) CreateMessage(ctx context.Context, session *conversation.Session, fromUserId, toUserId, content string) (*conversation.Message, error) {
	r, err := c.client.Message.
		Create().
		SetSession(toEntSession(session)).
		SetFromUserID(fromUserId).
		SetToUserID(toUserId).
		SetContent(content).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("Create Message failed: %w", err)
	}
	return toConversationMessage(r), nil
}

func (c *ConversationHandler) CreateSpouseMessage(ctx context.Context, session *conversation.Session, fromUserId, toUserId, content string, spouse *conversation.Message) (*conversation.Message, error) {
	r, err := c.client.Message.
		Create().
		SetSession(toEntSession(session)).
		SetFromUserID(fromUserId).
		SetToUserID(toUserId).
		SetContent(content).
		SetSpouse(toEntMessage(spouse)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("Create Spouse Message failed: %w", err)
	}
	return toConversationMessage(r), nil
}

func (c *ConversationHandler) ListLatestMessagesWithSpouse(ctx context.Context, session *conversation.Session, userId string, turns int) ([]*conversation.Message, error) {
	sessionId := session.ID
	msgs, err := c.client.Message.
		Query().
		Where(message.SessionIDEQ(sessionId), message.FromUserIDEQ(userId), message.HasSpouse()).
		Order(chatent.Desc(message.FieldCreatedAt)).
		Limit(turns).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query message failed: %w", err)
	}

	spouseMsgs, err := c.client.Message.
		Query().
		Where(message.SessionIDEQ(sessionId), message.ToUserIDEQ(userId), message.HasSpouse()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query spouse message failed: %w", err)
	}
	spouseMsgMap := make(map[int]*chatent.Message, 0)
	for _, m := range spouseMsgs {
		spouseMsgMap[m.SpouseID] = m
	}

	result := make([]*conversation.Message, 0)
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Sub(msgs[j].CreatedAt) < 0
	})
	for _, m := range msgs {
		spouse, ok := spouseMsgMap[m.ID]
		if ok {
			result = append(result, toConversationMessage(m), toConversationMessage(spouse))
		}
	}
	return result, nil
}

func toConversationSession(s *chatent.Session) *conversation.Session {
	return &conversation.Session{
		ID:        s.ID,
		UserID:    s.UserID,
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		DeletedAt: s.DeletedAt,
	}
}

func toEntSession(s *conversation.Session) *chatent.Session {
	return &chatent.Session{
		ID:        s.ID,
		UserID:    s.UserID,
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		DeletedAt: s.DeletedAt,
	}
}

func toConversationMessage(m *chatent.Message) *conversation.Message {
	return &conversation.Message{
		ID:         m.ID,
		SessionID:  m.SessionID,
		FromUserID: m.FromUserID,
		ToUserID:   m.ToUserID,
		Content:    m.Content,
		SpouseID:   m.SpouseID,
		CreatedAt:  m.CreatedAt,
	}
}

func toEntMessage(m *conversation.Message) *chatent.Message {
	return &chatent.Message{
		ID:         m.ID,
		SessionID:  m.SessionID,
		FromUserID: m.FromUserID,
		ToUserID:   m.ToUserID,
		Content:    m.Content,
		SpouseID:   m.SpouseID,
		CreatedAt:  m.CreatedAt,
	}
}
