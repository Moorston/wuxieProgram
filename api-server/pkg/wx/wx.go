package wx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	appID     string
	secret    string
	httpClient *http.Client

	accessToken string
	tokenExpiry time.Time
	mu          sync.Mutex
}

func NewClient(appID, secret string) *Client {
	return &Client{
		appID:      appID,
		secret:     secret,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) getAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", c.appID, c.secret)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wx api error: %d %s", result.ErrCode, result.ErrMsg)
	}

	c.accessToken = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	return c.accessToken, nil
}

type SubscribeMsgReq struct {
	ToUser     string            `json:"touser"`
	TemplateID string            `json:"template_id"`
	Page       string            `json:"page"`
	Data       map[string]TemplateData `json:"data"`
}

type TemplateData struct {
	Value string `json:"value"`
}

func (c *Client) SendSubscribeMessage(openID, templateID, page string, data map[string]string) error {
	token, err := c.getAccessToken()
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	msgData := make(map[string]TemplateData)
	for k, v := range data {
		msgData[k] = TemplateData{Value: v}
	}

	req := SubscribeMsgReq{
		ToUser:     openID,
		TemplateID: templateID,
		Page:       page,
		Data:       msgData,
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", token)

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("send subscribe message: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wx api error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}
