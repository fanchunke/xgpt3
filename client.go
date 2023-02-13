package xgpt3

import (
	"context"
	"fmt"

	"github.com/fanchunke/xgpt3/conversation"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-gpt3"
)

const (
	defaultMaxToken = 4000
	defaultMaxTurn  = 10
	defaultChannel  = "default"
)

type Client struct {
	*gogpt.Client
	ch       conversation.Handler
	maxToken int
	maxTurn  int
	logger   zerolog.Logger
}

func NewClient(client *gogpt.Client, ch conversation.Handler) *Client {
	return &Client{Client: client, ch: ch, maxToken: defaultMaxToken, maxTurn: defaultMaxTurn, logger: log.Logger}
}

func (c *Client) WithMaxTurn(n int) *Client {
	c.maxTurn = n
	return c
}

func (c *Client) WithLogger(l zerolog.Logger) *Client {
	c.logger = l
	return c
}

func (c *Client) CreateConversationCompletion(
	ctx context.Context,
	request gogpt.CompletionRequest,
) (gogpt.CompletionResponse, error) {
	return c.CreateConversationCompletionWithChannel(ctx, request, defaultChannel)
}

func (c *Client) CreateConversationCompletionWithChannel(
	ctx context.Context,
	request gogpt.CompletionRequest,
	channel string,
) (gogpt.CompletionResponse, error) {
	c.logger.Debug().Msgf("User: %s, Origin Prompt: %s", request.User, request.Prompt)
	// 预处理
	session, msg, err := c.preCompletion(ctx, &request, channel)
	if err != nil {
		return gogpt.CompletionResponse{}, fmt.Errorf("preprocess failed: %w", err)
	}
	c.logger.Debug().Msgf("User: %s, Prompt with conversation: %s", request.User, request.Prompt)

	// 请求
	resp, err := c.Client.CreateCompletion(ctx, request)
	if err != nil {
		return resp, err
	}

	// 后处理
	_, err = c.postCompletion(ctx, request, resp, session, msg, channel)
	if err != nil {
		return gogpt.CompletionResponse{}, fmt.Errorf("postprocess failed: %w", err)
	}
	return resp, nil
}

func (c *Client) preCompletion(ctx context.Context, request *gogpt.CompletionRequest, channel string) (*conversation.Session, *conversation.Message, error) {
	session, newPrompt := c.buildSessionQuery(ctx, request.User, request.Prompt)

	var err error
	// 如果没有 session，创建一个 session
	if session == nil {
		session, err = c.ch.CreateSession(ctx, request.User)
		if err != nil {
			return nil, nil, fmt.Errorf("create session failed: %w", err)
		}
	}

	// 保存用户消息
	msg, err := c.ch.CreateMessage(ctx, session, request.User, channel, request.Prompt)
	if err != nil {
		return session, nil, fmt.Errorf("create message failed: %w", err)
	}

	if len(newPrompt) > c.maxToken {
		newPrompt = string([]rune(newPrompt)[len(newPrompt)-c.maxToken:])
	}

	request.Prompt = newPrompt
	return session, msg, nil
}

func (c *Client) buildSessionQuery(ctx context.Context, userId, query string) (*conversation.Session, string) {
	session, err := c.ch.GetLatestActiveSession(ctx, userId)
	if err != nil {
		fmt.Println(err)
		return nil, query
	}
	msgs, err := c.ch.ListLatestMessagesWithSpouse(ctx, session, userId, c.maxTurn)
	if err != nil {
		return session, query
	}

	result := ""
	for _, m := range msgs {
		if m.FromUserID == userId {
			result += fmt.Sprintf("Q: %s\n", m.Content)
		} else {
			result += fmt.Sprintf("A: %s\n", m.Content)
		}
	}
	result += fmt.Sprintf("Q: %s\nA: ", query)
	return session, result
}

func (c *Client) postCompletion(ctx context.Context, request gogpt.CompletionRequest, response gogpt.CompletionResponse, session *conversation.Session, msg *conversation.Message, channel string) (*conversation.Message, error) {
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("Empty GPT Choices")
	}

	reply := response.Choices[0].Text
	m, err := c.ch.CreateSpouseMessage(ctx, session, channel, request.User, reply, msg)
	if err != nil {
		return nil, fmt.Errorf("create spouse message failed: %w", err)
	}
	return m, nil
}
