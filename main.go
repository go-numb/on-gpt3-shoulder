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
	fmt.Println(APITOKEN)
}

func main() {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Static("/", "static")

	client := &Client{
		c:   gogpt.NewClient(""),
		ctx: context.Background(),
	}

	api := e.Group("/api")

	api.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "ping")
	})
	api.POST("/", client.Request)

	e.Logger.Fatal(e.Start(":8080"))
}

type Body struct {
	Q string `json:"q"`
}

func (p *Client) Request(c echo.Context) error {
	t := time.Now()
	defer fmt.Printf("%fs\n", time.Since(t).Seconds())

	body := new(Body)
	if err := c.Bind(body); err != nil {
		return err
	}

	if body.Q == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"data": errors.New("has not query")})
	}

	req := gogpt.CompletionRequest{
		Prompt:      fmt.Sprintf("%s。", body.Q),
		Temperature: 0.6,
		// Model:            "text-ada-001",
		Model:            "text-davinci-003",
		MaxTokens:        400,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}

	data := gogpt.CompletionResponse{}
	for i := 0; i < 5; i++ {
		res, err := p.c.CreateCompletion(p.ctx, req)
		if err != nil {
			return err
		}

		if res.Choices[0].FinishReason == "length" {
			req.MaxTokens = int(float64(req.MaxTokens) * 1.5)
			time.Sleep(2 * time.Second)
			log.Printf("request times: %d", i+1)
			continue
		}

		data = res
		break
	}

	data.Choices[0].Text = strings.Replace(data.Choices[0].Text, "\n\nA: ", "", 1)
	data.Choices[0].Text = strings.Replace(data.Choices[0].Text, "A:", "", 1)

	return c.JSON(http.StatusOK, data)
}
