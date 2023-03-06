# xgpt3

[go-openai](https://github.com/sashabaranov/go-openai) API wrapper that supports conversation

## Installation

```bash
go get github.com/fanchunke/xgpt3
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fanchunke/xgpt3"
	"github.com/fanchunke/xgpt3/conversation/ent"
	"github.com/fanchunke/xgpt3/conversation/ent/chatent"
	"github.com/sashabaranov/go-openai"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	entClient, err := chatent.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/chatgpt?parseTime=True")
	if err != nil {
		log.Fatalf("Open database failed: %s", err)
	}

	// 生成数据库表
	if err := entClient.Schema.Create(context.Background()); err != nil {
		log.Fatalf("Create database schema failed: %s", err)
	}

	// conversation handler
	handler := ent.New(entClient)

	// openai client
	gptClient := openai.NewClient("authToken")

	// xgpt3 client
	xgpt3Client := xgpt3.NewClient(gptClient, handler)

	// 请求
	req := openai.CompletionRequest{
		Model:           openai.GPT3TextDavinci003,
		MaxTokens:       100,
		Prompt:          "Lorem ipsum",
		TopP:            1,
		Temperature:     0.9,
		PresencePenalty: 0.6,
		User:            "fanchunke",
	}
	resp, err := xgpt3Client.CreateConversationCompletion(context.Background(), req)
	if err != nil {
		log.Fatalf("CreateConversationCompletion failed: %s", err)
	}
	fmt.Println(resp.Choices[0].Text)
}
```