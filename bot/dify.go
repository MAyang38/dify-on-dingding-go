package bot

import (
	"bufio"
	"bytes"
	"ding/clients"
	"ding/consts"
	"ding/utils"
	"encoding/json"
	"errors"
	"fmt"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	"github.com/alibabacloud-go/tea/tea"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type difySession struct {
	conversationID string
	expiry         time.Time
}

type difyClient struct {
	ApiBase    string
	DifyApiKey string
	Sessions   map[string]difySession // 存储会话
	mu         sync.Mutex             // 保护 Sessions 免受并发访问问题
}

var DifyClient difyClient

func InitDifyClient() {
	API_KEY := os.Getenv("API_KEY")
	API_URL := os.Getenv("API_URL")
	DifyClient = difyClient{
		ApiBase:    API_URL,
		DifyApiKey: API_KEY,
		Sessions:   make(map[string]difySession),
	}
}

type RequestBody struct {
	Inputs         map[string]interface{} `json:"inputs"`
	Query          string                 `json:"query"`
	ResponseMode   string                 `json:"response_mode"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	User           string                 `json:"user,omitempty"`
}

type ApiResponse struct {
	Event          string                 `json:"event"`
	TaskID         string                 `json:"task_id"`
	ID             string                 `json:"id"`
	MessageID      string                 `json:"message_id"`
	ConversationID string                 `json:"conversation_id"`
	Mode           string                 `json:"mode"`
	Answer         string                 `json:"answer"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      int64                  `json:"created_at"`
}

type StreamingEvent struct {
	Event          string                 `json:"event"`
	TaskID         string                 `json:"task_id,omitempty"`
	WorkflowRunID  string                 `json:"workflow_run_id,omitempty"`
	MessageID      string                 `json:"message_id,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	ID             string                 `json:"id,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Answer         string                 `json:"answer,omitempty"`
	CreatedAt      int64                  `json:"created_at,omitempty"`
	FinishedAt     int64                  `json:"finished_at,omitempty"`
}

// 添加会话
func (client *difyClient) AddSession(userID, conversationID string) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.Sessions[userID] = difySession{
		conversationID: conversationID,
		expiry:         time.Now().Add(30 * time.Minute),
	}
}

// 获取会话
func (client *difyClient) GetSession(userID string) (string, bool) {
	client.mu.Lock()
	defer client.mu.Unlock()
	sess, exists := client.Sessions[userID]
	if !exists {
		return "", false
	}
	if time.Now().After(sess.expiry) {
		// 会话已过期
		delete(client.Sessions, userID)
		return "", false
	}
	return sess.conversationID, true
}

func (client *difyClient) CallAPIBlock(query, conversationID, userID string) (string, error) {

	// 构建请求体
	requestBody := RequestBody{
		Inputs:         make(map[string]interface{}),
		Query:          query,
		ResponseMode:   "blocking",
		ConversationID: conversationID,
		User:           userID,
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", client.ApiBase+"/chat-messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.DifyApiKey)

	// 发送请求
	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d, response: %s", resp.StatusCode, string(body))
	}
	var response ApiResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("【CallAPI】转换异常", err)
		return "", err
	}
	client.AddSession(userID, response.ConversationID)
	return response.Answer, nil
}

func (client *difyClient) CallAPIStreaming(query, conversationID, userID string, cardInstanceId string) error {
	// 初始化客户端
	clientHttp := &http.Client{}
	// 构建请求体
	requestBody := RequestBody{
		Inputs:         make(map[string]interface{}),
		Query:          query,
		ResponseMode:   "streaming",
		ConversationID: conversationID,
		User:           userID,
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	// 创建请求
	req, err := http.NewRequest("POST", client.ApiBase+"/chat-messages", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// 设置必要的请求头
	req.Header.Set("Authorization", "Bearer "+client.DifyApiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := clientHttp.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		// 读取响应体
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return err
		}
		bodyString := string(bodyBytes)

		// 打印错误信息
		fmt.Printf("Error: received non-200 response code: %d\n", resp.StatusCode)
		fmt.Printf("Response body: %s\n", bodyString)
		return errors.New("Error: received non-200 response code")
	}

	scanner := bufio.NewScanner(resp.Body)
	var answerBuilder strings.Builder
	cm := utils.NewChannelManager()
	defer func() {
		if !cm.IsClosed() {
			cm.CloseChannel()
		}
	}()
	//updateChannel := make(chan string, 1)
	//defer close(updateChannel)
	//stopChannel := make(chan struct{}) // 创建停止信号通道

	go func(*utils.ChannelManager) {
		var lastContent string
		timer := time.NewTicker(200 * time.Millisecond) // 每200ms触发一次
		defer timer.Stop()
		for {
			select {
			case content := <-cm.DataCh:
				{
					fmt.Println("接收到的内容", content)
					lastContent = content
				}
			case <-timer.C:
				if lastContent != "" {
					go func(content string) {
						err := updateDingTalkCard(content, cardInstanceId)
						if err != nil {
							fmt.Println("Error updating DingTalk card:", err)
						}
					}(lastContent)
					lastContent = ""
				}
			case <-cm.CloseCh: // 接收到停止信号，退出循环
				return
			}
		}
	}(cm)

	for scanner.Scan() {
		var event StreamingEvent
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
		}
		if err = json.Unmarshal([]byte(line), &event); err != nil {
			fmt.Println("Error decoding JSON:", err)
			continue
		}

		err = client.processEvent(userID, event, &answerBuilder, cm)
		if err != nil {
			return err
		}

	}

	if err = scanner.Err(); err != nil {
		fmt.Println("Error reading response:", err)
		return err
	}
	if !cm.IsClosed() {
		cm.CloseChannel()
	}
	fmt.Println("Final Answer:", answerBuilder.String())
	time.Sleep(300)
	err = updateDingTalkCard(answerBuilder.String(), cardInstanceId)
	if err != nil {
		fmt.Println("Error updating DingTalk card:", err)
	}

	return nil
}
func (client *difyClient) processEvent(userID string, event StreamingEvent, answerBuilder *strings.Builder, cm *utils.ChannelManager) error {
	switch event.Event {
	case "message":
		{
			answerBuilder.WriteString(event.Answer)
			select {
			case cm.DataCh <- answerBuilder.String():
				time.Sleep(10)
			default:
			}
		}
	case "message_end":
		{
			// 发送停止信号
			cm.CloseChannel()
			client.AddSession(userID, event.ConversationID)
		}
	case "message_replace":
		{

		}
	case "error":
		{
			// 发送停止信号
			cm.CloseChannel()
			return errors.New("dify err")
		}
	}
	return nil

}
func updateDingTalkCard(content string, cardInstanceId string) error {
	fmt.Println("发送内容:", content)

	timeStart := time.Now()
	title := ""
	updateRequest := &dingtalkim_1_0.UpdateRobotInteractiveCardRequest{
		CardBizId: tea.String(cardInstanceId),
		CardData:  tea.String(fmt.Sprintf(consts.MessageCardTemplate, title, content)),
	}
	_, err := clients.DingtalkClient1.UpdateInteractiveCard(updateRequest)
	if err != nil {
		return err
	}
	elapsed := time.Since(timeStart)
	fmt.Printf("updateDingTalkCard 执行时间: %s\n", elapsed)
	return nil
}