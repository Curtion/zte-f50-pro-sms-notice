package bark

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client Bark 客户端
type Client struct {
	keys   []string
	sound  string
	client *http.Client
}

// NewClient 创建新的 Bark 客户端
func NewClient(keys []string, sound string) *Client {
	return &Client{
		keys:   keys,
		sound:  sound,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Send 发送 Bark 通知
// title: 通知标题
// body: 通知内容
func (c *Client) Send(title, body string) error {
	if len(c.keys) == 0 {
		return fmt.Errorf("未配置 Bark key")
	}

	var lastErr error

	for _, key := range c.keys {
		if key == "" {
			continue
		}

		if err := c.sendToDevice(key, title, body); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// sendToDevice 向单个设备发送通知
func (c *Client) sendToDevice(key, title, body string) error {
	// Bark API URL 格式: https://api.day.app/{key}/{title}/{body}
	// 或者使用 POST 请求

	// 对标题和内容进行 URL 编码
	encodedTitle := url.QueryEscape(title)
	encodedBody := url.QueryEscape(body)

	// 构建 URL
	apiURL := fmt.Sprintf("https://api.day.app/%s/%s/%s", key, encodedTitle, encodedBody)

	// 添加参数
	params := url.Values{}
	params.Set("sound", c.sound)

	reqURL := apiURL + "?" + params.Encode()

	resp, err := c.client.Get(reqURL)
	if err != nil {
		return fmt.Errorf("发送 Bark 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bark API 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// SendWithOptions 发送带更多选项的 Bark 通知
func (c *Client) SendWithOptions(key, title, body string, options map[string]string) error {
	encodedTitle := url.QueryEscape(title)
	encodedBody := url.QueryEscape(body)

	apiURL := fmt.Sprintf("https://api.day.app/%s/%s/%s", key, encodedTitle, encodedBody)

	// 添加额外参数
	params := url.Values{}
	for k, v := range options {
		params.Set(k, v)
	}

	if c.sound != "" && params.Get("sound") == "" {
		params.Set("sound", c.sound)
	}

	reqURL := apiURL
	if len(params) > 0 {
		reqURL = apiURL + "?" + params.Encode()
	}

	resp, err := c.client.Get(reqURL)
	if err != nil {
		return fmt.Errorf("发送 Bark 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bark API 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// SendPost 使用 POST 方式发送通知（适合长内容）
func (c *Client) SendPost(key, title, body string) error {
	apiURL := fmt.Sprintf("https://api.day.app/%s", key)

	data := url.Values{
		"title": {title},
		"body":  {body},
		"sound": {c.sound},
	}

	resp, err := c.client.Post(apiURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("发送 Bark 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bark API 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}
