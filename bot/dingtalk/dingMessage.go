package dingbot

import (
	"bufio"
	"context"
	"ding/bot"
	"ding/clients"
	"ding/consts"
	"ding/models"
	selfutils "ding/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"strings"
	"sync"
	"time"
)

type DingMessage struct {
	Ctx              context.Context
	Data             *chatbot.BotCallbackDataModel
	MsgType          string
	Permission       int
	IsGroup          bool
	CardInstanceId   string
	ReceivedMsgStr   string
	ConversationID   string
	ImageCodeList    []string
	ImageUrlList     []string
	ProcessStartTime time.Time
	ProcessEndTime   time.Time
	ProcessDurTime   time.Duration
}

var (
	messageQueue    chan *DingMessage
	wg              sync.WaitGroup
	dingSupportType []string
)

func DingVarInit() {
	messageQueue = make(chan *DingMessage, 1000) // 设置队列容量
	dingSupportType = []string{"text", "audio", "picture"}
	wg.Add(1)
	go messageConsumer()
}

func messageConsumer() {
	defer wg.Done()
	for msg := range messageQueue {
		// 处理消息的逻辑
		msg.processMessage()
	}
}

func (msg *DingMessage) startProcessing() {
	msg.ProcessStartTime = time.Now()
}

func (msg *DingMessage) endProcessing() {
	msg.ProcessEndTime = time.Now()
	msg.ProcessDurTime = msg.ProcessEndTime.Sub(msg.ProcessStartTime)
	fmt.Println("Duration:", msg.ProcessDurTime)
}
func (msg *DingMessage) processMessage() {
	msg.startProcessing()
	if msg.ReceivedMsgStr != "" {
		// 获取用户sessionId
		userID := msg.Data.SenderId
		conversationID, exists := bot.DifyClient.GetSession(msg.Data.SenderId)
		if exists {
			fmt.Println("Conversation ID for user:", userID, "is", conversationID)
		} else {
			conversationID = ""
			fmt.Println("No conversation ID found for user:", userID)
		}
		msg.ConversationID = conversationID
		// 调用dify API 获取工作流
		difyResp, err := bot.DifyClient.CallAPIStreaming(msg.ReceivedMsgStr, userID, conversationID, msg.Permission)
		if err != nil {
			fmt.Println("Error CallAPIStreaming:", err)
			return
		}
		defer difyResp.Body.Close()
		// 发送卡片
		u, err := uuid.NewUUID()
		if err != nil {
			fmt.Println("生成uuid错误")
			return
		}
		cardInstanceId := u.String()
		msg.CardInstanceId = cardInstanceId
		// 接收流返回
		var answerBuilder strings.Builder
		cm := selfutils.NewChannelManager()
		defer func() {
			if !cm.IsClosed() {
				cm.CloseChannel()
			}
		}()

		go func(cm *selfutils.ChannelManager, cardInstanceId string) {
			var lastContent string
			timer := time.NewTicker(200 * time.Millisecond) // 每200ms触发一次
			defer timer.Stop()
			for {
				select {
				case content := <-cm.DataCh:
					{
						//fmt.Println("接收到的内容", content)
						lastContent = content
					}
				case <-timer.C:
					if lastContent != "" {
						go func(content string) {
							cardData := fmt.Sprintf(consts.MessageCardTemplateWithTitle1, content)
							err := UpdateDingTalkCard(cardData, cardInstanceId)
							if err != nil {
								fmt.Println("Error updating DingTalk card:", err)
							}
						}(lastContent)
						lastContent = ""
					}
				case <-cm.CloseCh: // 接收到停止信号，退出循环
					return
					//case <-cm.CardStart:
					//sendInteractiveCard(cardInstanceId, msg)
					//cm.CardStart = nil
				}

			}
		}(cm, cardInstanceId)
		sendInteractiveCard(cardInstanceId, msg)
		streamScanner := bufio.NewScanner(difyResp.Body)
		for streamScanner.Scan() {
			var event bot.StreamingEvent
			line := streamScanner.Text()
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

			err = bot.DifyClient.ProcessEvent(userID, event, &answerBuilder, cm)
			if err != nil {
				cardData := fmt.Sprintf(consts.MessageCardTemplateWithoutTitle, "服务器内部错误")
				err = UpdateDingTalkCard(cardData, cardInstanceId)
				fmt.Printf("processEvent err %s\n", err)
				return
			}

		}

		if err = streamScanner.Err(); err != nil {
			fmt.Println("Error reading response:", err)
			return
		}
		if !cm.IsClosed() {
			cm.CloseChannel()
		}
		fmt.Println("Final Answer:", answerBuilder.String())
		time.Sleep(300)
		cardData := fmt.Sprintf(consts.MessageCardTemplateWithoutTitle, answerBuilder.String())
		err = UpdateDingTalkCard(cardData, cardInstanceId)
		if err != nil {
			fmt.Println("Error updating DingTalk card:", err)
		}
		// 结束处理
		msg.endProcessing()
		// 记录发送日志
		if clients.PermissionControlInit == 1 {
			conversationID, _ := bot.DifyClient.GetSession(userID)
			msg.ConversationID = conversationID
			question := models.Question{
				Name:      msg.Data.SenderNick,
				Query:     msg.ReceivedMsgStr,
				Reply:     answerBuilder.String(),
				UserId:    userID,
				SessionId: msg.ConversationID,
			}
			if msg.IsGroup {
				question.ChatType = 2
			} else {
				question.ChatType = 1
			}

			err = clients.QuestionLogCli.SendQueryRecord(question)
			if err != nil {
				fmt.Println("添加问题日志出错", err)
				return
			}
		}
	}
}
