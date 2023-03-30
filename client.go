package xgpt3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fanchunke/xgpt3/conversation"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultMaxCtxLength = 4097
	defaultMaxTurn      = 10
	defaultChannel      = "default"
	questionPrefix      = "Q"
	answerPrefix        = "A"
)

type Client struct {
	*openai.Client
	ch           conversation.Handler
	maxCtxLength int
	maxTurn      int
	logger       zerolog.Logger
}

func NewClient(client *openai.Client, ch conversation.Handler) *Client {
	return &Client{Client: client, ch: ch, maxCtxLength: defaultMaxCtxLength, maxTurn: defaultMaxTurn, logger: log.Logger}
}

func (c *Client) WithMaxTurn(n int) *Client {
	c.maxTurn = n
	return c
}

func (c *Client) WithMaxCtxLength(n int) *Client {
	c.maxCtxLength = n
	return c
}

func (c *Client) WithLogger(l zerolog.Logger) *Client {
	c.logger = l
	return c
}

func (c *Client) CreateConversationCompletion(
	ctx context.Context,
	request openai.CompletionRequest,
) (openai.CompletionResponse, error) {
	return c.CreateConversationCompletionWithChannel(ctx, request, defaultChannel)
}

func (c *Client) CreateConversationCompletionWithChannel(
	ctx context.Context,
	request openai.CompletionRequest,
	channel string,
) (openai.CompletionResponse, error) {
	c.logger.Debug().Msgf("User: %s, Origin Prompt: %s", request.User, request.Prompt)
	// 预处理
	session, msg, err := c.preCompletion(ctx, &request, channel)
	if err != nil {
		return openai.CompletionResponse{}, fmt.Errorf("preprocess failed: %w", err)
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
		return openai.CompletionResponse{}, fmt.Errorf("postprocess failed: %w", err)
	}
	return resp, nil
}

func (c *Client) preCompletion(ctx context.Context, request *openai.CompletionRequest, channel string) (*conversation.Session, *conversation.Message, error) {
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

func (c *Client) buildSessionQuery(ctx context.Context, session *conversation.Session, request *openai.CompletionRequest) string {
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

func (c *Client) postCompletion(ctx context.Context, request openai.CompletionRequest, response openai.CompletionResponse, session *conversation.Session, msg *conversation.Message, channel string) (*conversation.Message, error) {
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

func getRequestTokens(request openai.ChatCompletionRequest) int {
	l := 0
	for _, m := range request.Messages {
		l += len(m.Content)
	}
	return l
}

func marshalMessages(msgs []openai.ChatCompletionMessage) string {
	s, _ := json.Marshal(msgs)
	return string(s)
}

func (c *Client) reduceRequestMessages(request openai.ChatCompletionRequest) []openai.ChatCompletionMessage {
	msgs := make([]openai.ChatCompletionMessage, 0)
	l := 0
	for _, m := range request.Messages {
		if l+len(m.Content)+request.MaxTokens <= c.maxCtxLength {
			msgs = append(msgs, m)
			l = len(m.Content)
		} else {
			break
		}
	}

	if len(msgs) == 0 && len(request.Messages) > 0 {
		m := request.Messages[0]
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: string([]rune(m.Content)[:c.maxCtxLength-request.MaxTokens]),
		})
	}
	return msgs
}

func (c *Client) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return c.CreateChatCompletionWithChannel(ctx, request, defaultChannel)
}

func (c *Client) CreateChatCompletionWithChannel(ctx context.Context, request openai.ChatCompletionRequest, channel string) (openai.ChatCompletionResponse, error) {
	c.logger.Debug().Msgf("User: %s, Origin Messages: %s", request.User, marshalMessages(request.Messages))
	// 预处理
	session, msg, err := c.preChatCompletion(ctx, &request, channel)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("chat completion preprocess failed: %w", err)
	}
	c.logger.Debug().Msgf("User: %s, Messages with conversation: %s", request.User, marshalMessages(request.Messages))

	// 请求
	resp, err := c.Client.CreateChatCompletion(ctx, request)
	if err != nil {
		return resp, err
	}

	// 后处理
	_, err = c.postChatCompletion(ctx, request, resp, session, msg, channel)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("chat completion postprocess failed: %w", err)
	}
	return resp, nil
}

func (c *Client) preChatCompletion(ctx context.Context, request *openai.ChatCompletionRequest, channel string) (*conversation.Session, *conversation.Message, error) {
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

	// 保存用户消息。只保存请求中最后一次的用户信息
	var msg *conversation.Message
	for i := len(request.Messages) - 1; i >= 0; i-- {
		m := request.Messages[i]
		if m.Role == openai.ChatMessageRoleUser {
			msg, err = c.ch.CreateMessage(ctx, session, request.User, channel, m.Content)
			if err != nil {
				return session, nil, fmt.Errorf("create message failed: %w", err)
			}
			break
		}
	}
	if msg == nil {
		return session, nil, errors.New("request has no user message")
	}

	// 构造会话历史消息
	request.Messages = c.buildChatSessionQuery(ctx, session, request)
	msgLen := getRequestTokens(*request)
	c.logger.Debug().Msgf("Requested %d tokens (%d in your messages; %d for the chat completion)", msgLen+request.MaxTokens, msgLen, request.MaxTokens)
	return session, msg, nil
}

func (c *Client) buildChatSessionQuery(ctx context.Context, session *conversation.Session, request *openai.ChatCompletionRequest) []openai.ChatCompletionMessage {
	msgLen := getRequestTokens(*request)
	if msgLen+request.MaxTokens > c.maxCtxLength {
		c.logger.Debug().Msgf("Requested %d tokens (%d in your messages; %d for the chat completion), reduce messages", msgLen+request.MaxTokens, msgLen, request.MaxTokens)
		return c.reduceRequestMessages(*request)
	}

	msgs, err := c.ch.ListLatestMessagesWithSpouse(ctx, session, request.User, c.maxTurn)
	if err != nil {
		c.logger.Warn().Msgf("ListLatestMessagesWithSpouse failed: %s", err)
		return request.Messages
	}

	// 按照消息创建时间倒序排序
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.Sub(msgs[j].CreatedAt) > 0
	})

	// 按照消息长度重排历史消息
	pl := msgLen
	selectedMsgs := append([]openai.ChatCompletionMessage{}, request.Messages...)
	for _, m := range msgs {
		var ccm openai.ChatCompletionMessage
		if m.FromUserID == request.User {
			ccm = openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: m.Content,
			}
		} else {
			ccm = openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: m.Content,
			}
		}

		if len(ccm.Content)+pl+request.MaxTokens <= c.maxCtxLength {
			selectedMsgs = append([]openai.ChatCompletionMessage{ccm}, selectedMsgs...)
			pl += len(ccm.Content)
		} else {
			break
		}
	}

	// 如果首个消息不是用户发出的，则忽略
	if len(selectedMsgs) > 0 && selectedMsgs[0].Role != openai.ChatMessageRoleUser {
		selectedMsgs = selectedMsgs[1:]
	}
	return selectedMsgs
}

func (c *Client) postChatCompletion(ctx context.Context, request openai.ChatCompletionRequest, response openai.ChatCompletionResponse, session *conversation.Session, msg *conversation.Message, channel string) (*conversation.Message, error) {
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("Empty GPT Choices")
	}

	reply := response.Choices[0].Message.Content
	m, err := c.ch.CreateSpouseMessage(ctx, session, channel, request.User, reply, msg)
	if err != nil {
		return nil, fmt.Errorf("create spouse message failed: %w", err)
	}
	return m, nil
}
