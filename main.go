package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"zte-sms-notice/bark"
	"zte-sms-notice/zte"
)

func main() {
	var (
		password = flag.String("p", "", "中兴 F50Pro 登录密码")
		barkKeys = flag.String("b", "", "Bark 通知 key，多设备用英文逗号分隔")
		sound    = flag.String("s", "healthnotification", "Bark 通知铃声名称")
		baseURL  = flag.String("url", "http://192.168.0.1", "中兴 F50Pro 地址")
		interval = flag.Int("i", 3, "检查短信间隔（秒）")
	)
	flag.Parse()

	if *password == "" {
		fmt.Println("错误: 必须提供密码参数 -p")
		flag.Usage()
		os.Exit(1)
	}

	if *barkKeys == "" {
		fmt.Println("错误: 必须提供 Bark key 参数 -b")
		flag.Usage()
		os.Exit(1)
	}

	// 解析多个 bark key
	keys := strings.Split(*barkKeys, ",")
	for i := range keys {
		keys[i] = strings.TrimSpace(keys[i])
	}

	log.Println("启动中兴 F50 Pro 短信监控...")
	log.Printf("设备地址: %s", *baseURL)
	log.Printf("检查间隔: %d 秒", *interval)

	// 创建 ZTE 客户端
	zteClient := zte.NewClient(*baseURL)

	// 创建 Bark 客户端
	barkClient := bark.NewClient(keys, *sound)

	// 记录已通知的短信 ID，避免重复通知
	notifiedIDs := make(map[string]bool)

	// 首次运行先登录
	if err := zteClient.Login(*password); err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	log.Println("登录成功")

	// 启动定时检查
	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	// 信号处理，优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 立即执行一次检查
	checkAndNotify(zteClient, barkClient, notifiedIDs)

	for {
		select {
		case <-ticker.C:
			checkAndNotify(zteClient, barkClient, notifiedIDs)
		case sig := <-sigChan:
			log.Printf("收到信号 %v，程序退出", sig)
			// 尝试登出
			if err := zteClient.Logout(); err != nil {
				log.Printf("登出失败: %v", err)
			} else {
				log.Println("已登出")
			}
			return
		}
	}
}

func checkAndNotify(zteClient *zte.Client, barkClient *bark.Client, notifiedIDs map[string]bool) {
	// 检查登录状态
	if err := zteClient.CheckLogin(); err != nil {
		log.Printf("登录状态检查失败: %v", err)
		return
	}

	// 获取短信列表（只获取未读的 tag=1）
	messages, err := zteClient.GetSMSList(0, 50, 1) // 1 = 只获取未读短信
	if err != nil {
		log.Printf("获取短信列表失败: %v", err)
		return
	}

	if len(messages) == 0 {
		return
	}

	log.Printf("发现 %d 条未读短信", len(messages))

	var newMessageIDs []string

	for _, msg := range messages {
		// 跳过已通知的短信
		if notifiedIDs[msg.ID] {
			continue
		}

		// 发送 Bark 通知
		title := fmt.Sprintf("来自 %s 的短信", msg.Number)
		if err := barkClient.Send(title, msg.Content); err != nil {
			log.Printf("发送 Bark 通知失败: %v", err)
			continue
		}

		log.Printf("已通知: [%s] %s", msg.Number, truncate(msg.Content, 30))
		notifiedIDs[msg.ID] = true
		newMessageIDs = append(newMessageIDs, msg.ID)
	}

	// 标记新通知的短信为已读
	if len(newMessageIDs) > 0 {
		if err := zteClient.MarkAsRead(newMessageIDs); err != nil {
			log.Printf("标记短信已读失败: %v", err)
		} else {
			log.Printf("已标记 %d 条短信为已读", len(newMessageIDs))
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
