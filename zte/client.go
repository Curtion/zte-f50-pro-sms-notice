package zte

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client ZTE 设备客户端
type Client struct {
	client  *http.Client
	baseURL string
	rd0     string
	rd1     string
	rd      string
	ad      string
}

// SMSMessage 短信消息
type SMSMessage struct {
	ID      string
	Number  string
	Content string
	Date    string
	IsNew   bool
}

// NewClient 创建新的 ZTE 客户端
func NewClient(baseURL string) *Client {
	return &Client{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// sha256Hex 计算 SHA256 并返回十六进制字符串
func sha256Hex(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// GetLD 获取 LD 参数（用于登录密码加密）
func (c *Client) GetLD() (string, error) {
	resp, err := c.client.Get(c.baseURL + "/goform/goform_get_cmd_process?cmd=LD&multi_data=1&isTest=false")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result["LD"], nil
}

// GetRD 获取 RD 参数
func (c *Client) GetRD() (string, error) {
	resp, err := c.client.Get(c.baseURL + "/goform/goform_get_cmd_process?cmd=RD&multi_data=1&isTest=false")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	c.rd = result["RD"]
	return c.rd, nil
}

// GetVersionInfo 获取版本信息（rd0 和 rd1）
func (c *Client) GetVersionInfo() error {
	resp, err := c.client.Get(c.baseURL + "/goform/goform_get_cmd_process?cmd=Language,cr_version,wa_inner_version&multi_data=1&isTest=false")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	c.rd0 = result["wa_inner_version"]
	c.rd1 = result["cr_version"]
	return nil
}

// GenerateAD 生成 AD 参数
func (c *Client) GenerateAD() string {
	if c.rd == "" || c.rd0 == "" || c.rd1 == "" {
		return ""
	}
	hash1 := sha256Hex(c.rd0 + c.rd1)
	c.ad = sha256Hex(hash1 + c.rd)
	return c.ad
}

// Login 登录设备
func (c *Client) Login(password string) error {
	// 获取 LD
	ld, err := c.GetLD()
	if err != nil {
		return fmt.Errorf("获取 LD 失败: %w", err)
	}

	// 加密密码: SHA256(SHA256(密码) + LD)
	hash1 := strings.ToUpper(sha256Hex(password))
	encryptedPass := strings.ToUpper(sha256Hex(hash1 + ld))

	data := url.Values{
		"isTest":   {"false"},
		"goformId": {"LOGIN"},
		"password": {encryptedPass},
	}

	resp, err := c.client.Post(c.baseURL+"/goform/goform_set_cmd_process",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		Result int `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	// result: 0 = 成功, 1 = 失败, 4 = 已登录, 5 = 已在别处登录
	if result.Result != 0 && result.Result != 4 {
		return fmt.Errorf("登录失败, 错误码: %d", result.Result)
	}

	// 获取版本信息和 RD，然后生成 AD
	if err := c.GetVersionInfo(); err != nil {
		return fmt.Errorf("获取版本信息失败: %w", err)
	}

	if _, err := c.GetRD(); err != nil {
		return fmt.Errorf("获取 RD 失败: %w", err)
	}

	c.GenerateAD()
	return nil
}

// CheckLogin 检查登录状态
func (c *Client) CheckLogin() error {
	resp, err := c.client.Get(c.baseURL + "/goform/goform_get_cmd_process?cmd=loginfo&multi_data=1&isTest=false")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result["loginfo"] != "ok" {
		return fmt.Errorf("未登录或登录已过期")
	}
	return nil
}

// Logout 登出
func (c *Client) Logout() error {
	if c.ad == "" {
		return nil
	}

	data := url.Values{
		"isTest":   {"false"},
		"goformId": {"LOGOUT"},
		"AD":       {c.ad},
	}

	resp, err := c.client.Post(c.baseURL+"/goform/goform_set_cmd_process",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetSMSList 获取短信列表
// memStore: 0=设备, 1=SIM卡
// tags: 0=全部, 1=未读, 2=已读, 3=发送, 4=草稿
func (c *Client) GetSMSList(page, pageSize, tags int) ([]SMSMessage, error) {
	// mem_store=1 表示 SIM 卡存储
	reqURL := fmt.Sprintf("%s/goform/goform_get_cmd_process?cmd=sms_data_total&page=%d&data_per_page=%d&mem_store=1&tags=%d&order_by=order+by+id+desc&isTest=false",
		c.baseURL, page, pageSize, tags)

	resp, err := c.client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Messages []struct {
			ID       string `json:"id"`
			Number   string `json:"number"`
			Content  string `json:"content"`
			Date     string `json:"date"`
			Tag      string `json:"tag"`
			GroupID  string `json:"draft_group_id"`
			Received string `json:"received_all_concat_sms"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var messages []SMSMessage
	for _, m := range result.Messages {
		// Base64 解码短信内容
		content, err := base64.StdEncoding.DecodeString(m.Content)
		if err != nil {
			content = []byte(m.Content) // 如果解码失败，使用原始内容
		}

		messages = append(messages, SMSMessage{
			ID:      m.ID,
			Number:  m.Number,
			Content: string(content),
			Date:    m.Date,
			IsNew:   m.Tag == "1",
		})
	}
	return messages, nil
}

// MarkAsRead 标记短信为已读
func (c *Client) MarkAsRead(ids []string) error {
	if c.ad == "" {
		return fmt.Errorf("AD 参数未生成")
	}

	msgID := strings.Join(ids, ";")
	if len(ids) > 0 {
		msgID += ";"
	}

	data := url.Values{
		"isTest":   {"false"},
		"goformId": {"SET_MSG_READ"},
		"msg_id":   {msgID},
		"tag":      {"0"},
		"AD":       {c.ad},
	}

	resp, err := c.client.Post(c.baseURL+"/goform/goform_set_cmd_process",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result["result"] != "success" {
		return fmt.Errorf("标记已读失败: %s", result["result"])
	}
	return nil
}

// DeleteSMS 删除短信
func (c *Client) DeleteSMS(ids []string) error {
	if c.ad == "" {
		return fmt.Errorf("AD 参数未生成")
	}

	msgID := strings.Join(ids, ";") + ";"

	data := url.Values{
		"isTest":      {"false"},
		"goformId":    {"DELETE_SMS"},
		"notCallback": {"true"},
		"msg_id":      {msgID},
		"AD":          {c.ad},
	}

	resp, err := c.client.Post(c.baseURL+"/goform/goform_set_cmd_process",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result["result"] != "success" {
		return fmt.Errorf("删除失败: %s", result["result"])
	}
	return nil
}
