package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type WiseRateResponse struct {
	Source string  `json:"source"`
	Target string  `json:"target"`
	Rate   float64 `json:"value"`
	Time   float64 `json:"time"`
}

type LarkMessage struct {
	MsgType string      `json:"msg_type"`
	Content LarkContent `json:"content"`
}

type LarkContent struct {
	Text string `json:"text"`
}

func getWiseRates() ([]WiseRateResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 需要获取的汇率对
	currencyPairs := []struct {
		source string
		target string
	}{
		{"USD", "CNY"},
		{"MYR", "CNY"},
		{"MYR", "HKD"},
	}

	var rates []WiseRateResponse

	for _, pair := range currencyPairs {
		url := fmt.Sprintf("https://wise.com/rates/live?source=%s&target=%s", pair.source, pair.target)

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("获取汇率失败 %s-%s: %v", pair.source, pair.target, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败 %s-%s: %v", pair.source, pair.target, err)
		}

		var rateData WiseRateResponse
		if err := json.Unmarshal(body, &rateData); err != nil {
			return nil, fmt.Errorf("解析JSON失败 %s-%s: %v", pair.source, pair.target, err)
		}

		rates = append(rates, rateData)
	}

	return rates, nil
}

func formatCurrencyName(currency string) string {
	names := map[string]string{
		"USD": "美金",
		"MYR": "马币",
		"CNY": "人民币",
		"HKD": "港币",
	}
	if name, exists := names[currency]; exists {
		return name
	}
	return currency
}

func sendLarkNotification(webhookURL string, rates []WiseRateResponse, previousRates []WiseRateResponse) error {
	// 格式化消息
	var messageBuffer bytes.Buffer

	for _, rate := range rates {
		sourceName := formatCurrencyName(rate.Source)
		
		// 查找上一次的汇率进行比较
		var arrow string
		if previousRates != nil {
			for _, prevRate := range previousRates {
				if rate.Source == prevRate.Source && rate.Target == prevRate.Target {
					if rate.Rate > prevRate.Rate {
						arrow = " ↑"
					} else if rate.Rate < prevRate.Rate {
						arrow = " ↓"
					} else {
						arrow = " →"
					}
					break
				}
			}
		} else {
			// 第一次运行时默认显示上升箭头
			arrow = " ↑"
		}
		
		messageBuffer.WriteString(fmt.Sprintf("%s%s-%s, 结汇: %.6f%s\n",
			sourceName, rate.Source, rate.Target, rate.Rate, arrow))
	}

	// 添加时间戳
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	messageBuffer.WriteString(fmt.Sprintf("更新时间: %s", timestamp))

	// 构造 Lark 消息
	message := LarkMessage{
		MsgType: "text",
		Content: LarkContent{
			Text: messageBuffer.String(),
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("构造消息失败: %v", err)
	}

	// 发送请求
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Lark API 返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

func startScheduler() {
	// 计算下一次整点的时间
	now := time.Now()
	nextHour := now.Truncate(time.Hour).Add(time.Hour)
	
	// 如果当前正好是整点，等待下一小时
	if nextHour.Before(now) || nextHour.Equal(now) {
		nextHour = nextHour.Add(time.Hour)
	}
	
	// 计算等待时间
	waitDuration := nextHour.Sub(now)
	log.Printf("等待 %v 后开始第一次执行，将在 %s 执行", waitDuration, nextHour.Format("2006-01-02 15:04:05"))
	
	// 等待到下一个整点
	time.Sleep(waitDuration)
	
	// 启动定时任务
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	// 每小时执行一次
	var previousRates []WiseRateResponse
	for t := range ticker.C {
		log.Printf("执行汇率通知任务 - %s", t.Format("2006-01-02 15:04:05"))
		
		// 获取当前汇率
		currentRates, err := getWiseRates()
		if err != nil {
			log.Printf("获取汇率失败: %v", err)
			continue
		}
		
		// 发送通知，使用上一次的汇率作为比较基准
		webhookURL := os.Getenv("LARK_WEBHOOK_URL")
		if webhookURL != "" {
			err = sendLarkNotification(webhookURL, currentRates, previousRates)
			if err != nil {
				log.Printf("发送通知失败: %v", err)
			} else {
				log.Println("汇率通知发送成功")
			}
		} else {
			log.Println("错误: 请设置环境变量 LARK_WEBHOOK_URL")
		}
		
		// 保存当前汇率作为下次比较的基准
		previousRates = currentRates
	}
}



func main() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Println("未找到 .env 文件，将使用系统环境变量")
	}

	log.Println("启动汇率通知服务...")
	startScheduler()
}
