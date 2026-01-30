# 中兴随身WiFi (F50 Pro) HTTP API 接口文档

## 基础信息

- **设备型号**: F50 Pro
- **基础URL**: `http://192.168.55.1`
- **数据格式**: JSON
- **编码**: UTF-8

---

## 1. 认证与会话管理

### 1.1 获取 LD 参数（用于登录密码加密）

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=LD&multi_data=1&isTest=false
```

**响应:**
```json
{
  "LD": "a1b2c3d4e5f6..."  // 随机字符串，用于密码加密
}
```

### 1.2 登录

**请求:**
```
POST /goform/goform_set_cmd_process
Content-Type: application/x-www-form-urlencoded
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| isTest | string | "false" |
| goformId | string | "LOGIN" |
| password | string | SHA256(SHA256(密码) + LD) |

**密码加密算法 (Go):**
```go
func encryptPassword(password, ld string) string {
    // 第一步: SHA256(密码)
    hash1 := sha256.Sum256([]byte(password))
    hash1Hex := hex.EncodeToString(hash1[:])
    
    // 第二步: SHA256(第一步结果 + LD)
    hash2 := sha256.Sum256([]byte(hash1Hex + ld))
    return hex.EncodeToString(hash2[:])
}
```

**响应:**
```json
{
  "result": 0  // 0 = 成功, 1 = 失败, 5 = 已在别处登录
}
```

### 1.3 获取 rd0, rd1 和 RD（用于生成 AD 参数）

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=RD&multi_data=1&isTest=false
```

**响应:**
```json
{
  "RD": "random_string_32chars"
}
```

**rd0 和 rd1** 从语言配置接口获取：

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=Language,cr_version,wa_inner_version&multi_data=1&isTest=false
```

**响应:**
```json
{
  "Language": "zh-cn",
  "wa_inner_version": "BD_xxxxxxx",  // rd0
  "cr_version": "1.0.0"              // rd1
}
```

**映射关系:**
- `rd0` = `wa_inner_version` (固件内部版本号)
- `rd1` = `cr_version` (定制版本/软件版本号)

### 1.4 生成 AD 参数

AD 参数用于除登录外的所有 POST 请求认证：

```go
func generateAD(rd0, rd1, rd string) string {
    // 第一步: SHA256(rd0 + rd1)
    hash1 := sha256.Sum256([]byte(rd0 + rd1))
    hash1Hex := hex.EncodeToString(hash1[:])
    
    // 第二步: SHA256(第一步结果 + RD)
    hash2 := sha256.Sum256([]byte(hash1Hex + rd))
    return hex.EncodeToString(hash2[:])
}
```

### 1.5 检查登录状态

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=loginfo&multi_data=1&isTest=false
```

**响应:**
```json
{
  "loginfo": "ok"  // "ok" = 已登录, 其他 = 未登录
}
```

### 1.6 登出

**请求:**
```
POST /goform/goform_set_cmd_process
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| isTest | string | "false" |
| goformId | string | "LOGOUT" |
| AD | string | 认证参数 |

---

## 2. 短信接口

### 2.1 获取短信容量信息

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=sms_capacity_info&isTest=false
```

**响应:**
```json
{
  "sms_nv_total": "100",           // 设备总容量
  "sms_nv_rev_total": "10",        // 设备收件箱数量
  "sms_nv_send_total": "5",        // 设备发件箱数量
  "sms_nv_draftbox_total": "2",    // 设备草稿箱数量
  "sms_sim_total": "50",           // SIM卡总容量
  "sms_sim_rev_total": "20",       // SIM卡收件箱数量
  "sms_sim_send_total": "0",       // SIM卡发件箱数量
  "sms_sim_draftbox_total": "0"    // SIM卡草稿箱数量
}
```

### 2.2 获取短信列表

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=sms_data_total&page=0&data_per_page=20&mem_store=0&tags=0&order_by=order+by+id+desc&isTest=false
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| cmd | string | "sms_data_total" |
| page | int | 页码，从0开始 |
| data_per_page | int | 每页数量 |
| mem_store | int | 存储位置: 0=设备, 1=SIM卡 |
| tags | int | 标签筛选: 10=全部, 1=未读 |
| order_by | string | 排序方式 |
| isTest | string | "false" |

**响应:**
```json
{
  "messages": [
    {
      "id": "12345",
      "number": "13800138000",           // 发送方号码
      "content": "5L2g5aW9",             // Base64编码的短信内容
      "date": "24,12,25,10,30,00",       // 日期: YY,MM,DD,HH,MM,SS
      "tag": "0",                        // 0=已读, 1=未读, 2=已发送
      "draft_group_id": "0",
      "received_all_concat_sms": "1"     // 1=完整长短信
    }
  ]
}
```

### 2.3 发送短信

**请求:**
```
POST /goform/goform_set_cmd_process
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| isTest | string | "false" |
| goformId | string | "SEND_SMS" |
| notCallback | string | "true" |
| Number | string | 接收方号码 |
| sms_time | string | 时间格式: "YY;MM;DD;HH;MM;SS;+8" |
| MessageBody | string | 编码后的短信内容（见下方编码说明） |
| ID | string | 消息ID（随机生成） |
| encode_type | string | 编码类型: "GSM7_default" 或 "UNICODE" |
| AD | string | 认证参数 |

**短信内容编码 (Go):**
```go
func encodeMessage(content string) string {
    var result strings.Builder
    for _, r := range content {
        // 将每个字符转为4位16进制
        result.WriteString(fmt.Sprintf("%04X", r))
    }
    return result.String()
}

func getEncodeType(content string) string {
    // GSM7 字符表（简化版）
    for _, r := range content {
        if r > 0x7F {
            return "UNICODE"
        }
    }
    return "GSM7_default"
}

func getCurrentTimeString() string {
    now := time.Now()
    tzOffset := now.Format("-07") // 时区偏移
    if now.Format("-07")[0] == '+' {
        tzOffset = now.Format("-07")
    } else {
        tzOffset = "+" + now.Format("07")
    }
    return fmt.Sprintf("%02d;%02d;%02d;%02d;%02d;%02d;%s",
        now.Year()%100, now.Month(), now.Day(),
        now.Hour(), now.Minute(), now.Second(),
        tzOffset)
}
```

**响应:**
```json
{
  "result": "success"  // 或 "failure"
}
```

### 2.4 删除单条/多条短信

**请求:**
```
POST /goform/goform_set_cmd_process
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| isTest | string | "false" |
| goformId | string | "DELETE_SMS" |
| notCallback | string | "true" |
| msg_id | string | 短信ID，多个用";"分隔，如 "123;456;" |
| AD | string | 认证参数 |

**响应:**
```json
{
  "result": "success"
}
```

### 2.5 删除全部短信

**请求:**
```
POST /goform/goform_set_cmd_process
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| isTest | string | "false" |
| goformId | string | "ALL_DELETE_SMS" |
| notCallback | string | "true" |
| which_cgi | string | "0"=设备, "1"=SIM卡 |
| AD | string | 认证参数 |

### 2.6 标记短信为已读

**请求:**
```
POST /goform/goform_set_cmd_process
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| isTest | string | "false" |
| goformId | string | "SET_MSG_READ" |
| msg_id | string | 短信ID，多个用";"分隔 |
| tag | string | "0" |
| AD | string | 认证参数 |

### 2.7 保存草稿

**请求:**
```
POST /goform/goform_set_cmd_process
```

**参数:**
| 字段 | 类型 | 说明 |
|------|------|------|
| isTest | string | "false" |
| goformId | string | "SAVE_SMS" |
| notCallback | string | "true" |
| SMSMessage | string | 编码后的内容 |
| SMSNumber | string | 号码，多个用";"分隔，以";"结尾 |
| Index | string | 草稿ID（新建为-1） |
| encode_type | string | "GSM7_default" 或 "UNICODE" |
| sms_time | string | 时间字符串 |
| draft_group_id | string | 草稿组ID |
| AD | string | 认证参数 |

### 2.8 获取短信发送报告

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=sms_status_rpt_data&page=0&data_per_page=20&isTest=false
```

**响应:**
```json
{
  "messages": [
    {
      "id": "12345",
      "number": "13800138000",
      "content": "delivery_report_content",
      "date": "24,12,25,10,30,00",
      "tag": "0"
    }
  ]
}
```

### 2.9 检查短信模块就绪状态

**请求:**
```
GET /goform/goform_get_cmd_process?cmd=sms_cmd_status_info&sms_cmd=1&isTest=false
```

**响应:**
```json
{
  "sms_cmd": "1",
  "sms_cmd_status_result": "3"  // "3" = 就绪
}
```

---

## 3. 完整的 Golang 示例代码

```go
package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"crypto/sha256"
	"time"
)

const (
	BaseURL = "http://192.168.55.1"
)

type ZTEClient struct {
	client *http.Client
	rd0    string
	rd1    string
	rd     string
	ad     string
	baseURL string
}

func NewZTEClient() *ZTEClient {
	return &ZTEClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: BaseURL,
		rd0:     "",  // 需要从语言配置获取
		rd1:     "",  // 需要从语言配置获取
	}
}

// SHA256 加密
func sha256Hex(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// 获取LD参数
func (c *ZTEClient) GetLD() (string, error) {
	resp, err := c.client.Get(c.baseURL + "/goform/goform_get_cmd_process?cmd=LD&multi_data=1&isTest=false")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result["LD"], nil
}

// 获取RD参数
func (c *ZTEClient) GetRD() (string, error) {
	resp, err := c.client.Get(c.baseURL + "/goform/goform_get_cmd_process?cmd=RD&multi_data=1&isTest=false")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	c.rd = result["RD"]
	return c.rd, nil
}

// 生成AD参数
func (c *ZTEClient) GenerateAD() string {
	if c.rd == "" || c.rd0 == "" || c.rd1 == "" {
		return ""
	}
	hash1 := sha256Hex(c.rd0 + c.rd1)
	c.ad = sha256Hex(hash1 + c.rd)
	return c.ad
}

// 登录
func (c *ZTEClient) Login(password string) error {
	ld, err := c.GetLD()
	if err != nil {
		return err
	}

	// 加密密码
	hash1 := sha256Hex(password)
	encryptedPass := sha256Hex(hash1 + ld)

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

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result["result"] != "0" && result["result"] != "4" {
		return fmt.Errorf("登录失败: %s", result["result"])
	}

	// 获取RD并生成AD
	c.GetRD()
	c.GenerateAD()
	return nil
}

// 获取短信列表
func (c *ZTEClient) GetSMSList(page, pageSize int) ([]SMSMessage, error) {
	url := fmt.Sprintf("%s/goform/goform_get_cmd_process?cmd=sms_data_total&page=%d&data_per_page=%d&mem_store=0&tags=0&order_by=order+by+id+desc&isTest=false",
		c.baseURL, page, pageSize)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
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
		content, _ := base64.StdEncoding.DecodeString(m.Content)
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

// 发送短信
func (c *ZTEClient) SendSMS(phoneNumber, content string) error {
	smsTime := getCurrentTimeString()
	encodedContent := encodeMessage(content)
	encodeType := getEncodeType(content)
	msgID := fmt.Sprintf("%d", time.Now().Unix())

	data := url.Values{
		"isTest":       {"false"},
		"goformId":     {"SEND_SMS"},
		"notCallback":  {"true"},
		"Number":       {phoneNumber},
		"sms_time":     {smsTime},
		"MessageBody":  {encodedContent},
		"ID":           {msgID},
		"encode_type":  {encodeType},
		"AD":           {c.ad},
	}

	resp, err := c.client.Post(c.baseURL+"/goform/goform_set_cmd_process",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result["result"] != "success" {
		return fmt.Errorf("发送失败: %s", result["result"])
	}
	return nil
}

// 删除短信
func (c *ZTEClient) DeleteSMS(ids []string) error {
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

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result["result"] != "success" {
		return fmt.Errorf("删除失败: %s", result["result"])
	}
	return nil
}

// 标记已读
func (c *ZTEClient) MarkAsRead(ids []string) error {
	msgID := strings.Join(ids, ";")
	if len(ids) > 0 {
		msgID += ";"
	}
	
	data := url.Values{
		"isTest":    {"false"},
		"goformId":  {"SET_MSG_READ"},
		"msg_id":    {msgID},
		"tag":       {"0"},
		"AD":        {c.ad},
	}

	_, err := c.client.Post(c.baseURL+"/goform/goform_set_cmd_process",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()))
	return err
}

// 工具函数
func getCurrentTimeString() string {
	now := time.Now()
	_, offset := now.Zone()
	tzHours := -offset / 3600
	return fmt.Sprintf("%02d;%02d;%02d;%02d;%02d;%02d;%+d",
		now.Year()%100, now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		tzHours)
}

func encodeMessage(content string) string {
	var result strings.Builder
	for _, r := range content {
		result.WriteString(fmt.Sprintf("%04X", r))
	}
	return result.String()
}

func getEncodeType(content string) string {
	for _, r := range content {
		if r > 0x7F {
			return "UNICODE"
		}
	}
	return "GSM7_default"
}

type SMSMessage struct {
	ID      string
	Number  string
	Content string
	Date    string
	IsNew   bool
}

// 使用示例
func main() {
	client := NewZTEClient()
	
	// 设置 rd0 和 rd1（从设备语言配置获取）
	client.rd0 = "your_rd0_here"
	client.rd1 = "your_rd1_here"
	
	// 登录
	if err := client.Login("admin"); err != nil {
		fmt.Println("登录失败:", err)
		return
	}
	fmt.Println("登录成功")
	
	// 获取短信列表
	messages, err := client.GetSMSList(0, 20)
	if err != nil {
		fmt.Println("获取短信失败:", err)
		return
	}
	
	for _, m := range messages {
		fmt.Printf("ID: %s, 来自: %s, 内容: %s\n", m.ID, m.Number, m.Content)
	}
	
	// 发送短信
	if err := client.SendSMS("13800138000", "Hello from Go!"); err != nil {
		fmt.Println("发送失败:", err)
	} else {
		fmt.Println("发送成功")
	}
}
```

---

## 4. 关键常量与配置

### 4.1 短信存储位置
- `0` - 设备存储
- `1` - SIM卡存储

### 4.2 短信标签类型
- `0` - 全部/已读
- `1` - 未读
- `2` - 已发送
- `3` - 草稿

### 4.3 编码类型
- `GSM7_default` - 英文/数字（7位编码）
- `UNICODE` - 中文/特殊字符（Unicode编码）

---

## 5. 错误处理

常见错误码:
- `result: 0` - 登录成功
- `result: 4` - 登录成功（另一种状态）
- `result: 1` - 登录失败
- `result: 2` - 重复登录（用户已在其他地方登录）
- `result: 3` - 密码错误
- `result: 5` - 已在LCD端登录
- `result: "failure"` - 操作失败

---

## 6. 注意事项

1. **会话保持**: 默认会话超时约5分钟，建议定期调用 `cmd=loginfo` 保持会话
2. **AD参数**: 每次获取RD后需要重新生成AD
3. **短信编码**: 中文内容必须使用UNICODE编码，并转为4位16进制字符串
4. **时区处理**: 短信时间需要包含时区偏移信息
5. **并发限制**: 避免同时发送多个请求，设备处理能力有限
