package xgpt3

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/fanchunke/xgpt3/conversation"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-gpt3"
)

const (
	defaultMaxCtxLength = 4097
	defaultMaxTurn      = 10
	defaultChannel      = "default"
	questionPrefix      = "Q"
	answerPrefix        = "A"
)

type Client struct {
	*gogpt.Client
	ch           conversation.Handler
	maxCtxLength int
	maxTurn      int
	logger       zerolog.Logger
}

func NewClient(client *gogpt.Client, ch conversation.Handler) *Client {
	return &Client{Client: client, ch: ch, maxCtxLength: defaultMaxCtxLength, maxTurn: defaultMaxTurn, logger: log.Logger}
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
	if request.MaxTokens >= c.maxCtxLength {
		return nil, nil, fmt.Errorf("request.MaxTokens exceeded maximum context length")
	}

	// 获取最近的 session。如果没有 session，创建一个 session。
	session, err := c.ch.GetLatestActiveSession(ctx, request.User)
	if err != nil {
		session, err = c.ch.CreateSession(ctx, request.User)
		if err != nil {
			return nil, nil, fmt.Errorf("create session failed: %w", err)
		}
	}

	newPrompt := c.buildSessionQuery(ctx, session, request)

	// 保存用户消息
	msg, err := c.ch.CreateMessage(ctx, session, request.User, channel, request.Prompt)
	if err != nil {
		return session, nil, fmt.Errorf("create message failed: %w", err)
	}

	request.Prompt = newPrompt
	c.logger.Debug().Msgf("Requested %d tokens (%d in your prompt; %d for the completion)", len(request.Prompt)+request.MaxTokens, len(request.Prompt), request.MaxTokens)
	return session, msg, nil
}

func (c *Client) buildSessionQuery(ctx context.Context, session *conversation.Session, request *gogpt.CompletionRequest) string {
	if len(request.Prompt)+request.MaxTokens > c.maxCtxLength {
		c.logger.Debug().Msgf("Requested %d tokens (%d in your prompt; %d for the completion), reduce prompt", len(request.Prompt)+request.MaxTokens, len(request.Prompt), request.MaxTokens)
		return string([]rune(request.Prompt)[:c.maxCtxLength-request.MaxTokens])
	}

	msgs, err := c.ch.ListLatestMessagesWithSpouse(ctx, session, request.User, c.maxTurn)
	if err != nil {
		c.logger.Warn().Msgf("ListLatestMessagesWithSpouse failed: %s", err)
		return request.Prompt
	}

	// 按照消息创建时间倒序排序
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Sub(msgs[j].CreatedAt) > 0
	})

	query := fmt.Sprintf("%s: %s\n%s: ", questionPrefix, request.Prompt, answerPrefix)
	pl := len(query)
	selectedMsgs := []string{query}
	for i := 0; i < len(msgs); i++ {
		m := msgs[i]
		prompt := ""
		if m.FromUserID == request.User {
			prompt += fmt.Sprintf("%s: %s\n", questionPrefix, m.Content)
		} else {
			prompt += fmt.Sprintf("%s: %s\n", answerPrefix, m.Content)
		}

		if len(prompt)+pl+request.MaxTokens <= c.maxCtxLength {
			selectedMsgs = append([]string{prompt}, selectedMsgs...)
			pl += len(prompt)
		} else {
			break
		}
	}

	result := ""
	for i, p := range selectedMsgs {
		if i == 0 && strings.HasPrefix(p, answerPrefix) {
			continue
		}
		result += p
	}

	return result
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

func (c *Client) CloseConversation(ctx context.Context, userId string) error {
	return c.ch.CloseSession(ctx, userId)
}
