package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gogpt "github.com/sashabaranov/go-gpt3"
)

type Client struct {
	c   *gogpt.Client
	ctx context.Context
}

var (
	APITOKEN = ""
)

func init() {
	f, err := ioutil.ReadFile("config.conf")
	if err != nil {
		log.Fatal(err)
	}

	APITOKEN = string(f)
}

func main() {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Static("/", "static")

	client := &Client{
		c:   gogpt.NewClient(APITOKEN),
		ctx: context.Background(),
	}

	api := e.Group("/api")

	api.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "pong")
	})
	api.POST("/", client.Request)

	e.Logger.Fatal(e.Start(":8080"))
}

type Body struct {
	Q string `json:"q"`
}

var (
	chats []gogpt.ChatCompletionMessage
)

func (p *Client) Request(c echo.Context) error {
	t := time.Now()
	defer fmt.Println(time.Since(t))

	body := new(Body)
	if err := c.Bind(body); err != nil {
		return err
	}

	fmt.Println(body)

	if body.Q == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"data": errors.New("has not query")})
	}

	req := gogpt.ChatCompletionRequest{
		// Model:            "text-ada-001",
		Model: gogpt.GPT3Dot5Turbo0301,
		Messages: append(chats, gogpt.ChatCompletionMessage{
			Role:    "user",
			Content: fmt.Sprintf("%s。", body.Q),
		}),
	}

	res, err := p.c.CreateChatCompletion(p.ctx, req)
	if err != nil {
		return err
	}

	res.Choices[0].Message.Content = strings.Replace(res.Choices[0].Message.Content, "\n\nA: ", "", 1)
	res.Choices[0].Message.Content = strings.Replace(res.Choices[0].Message.Content, "A:", "", 1)

	chats = append(chats, gogpt.ChatCompletionMessage{
		Role:    "user",
		Content: body.Q,
	})
	chats = append(chats, gogpt.ChatCompletionMessage{
		Role:    "assistant",
		Content: res.Choices[0].Message.Content,
	})

	fmt.Printf("%#v\n", res)

	return c.JSON(http.StatusOK, res)
}
