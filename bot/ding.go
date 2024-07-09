package bot

import (
	"context"
	"ding/clients"
	"ding/consts"
	"ding/models"
	selfutils "ding/utils"
	"encoding/json"
	"fmt"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	"github.com/alibabacloud-go/dingtalk/robot_1_0"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/utils"
	"os"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Ctx            context.Context
	Data           *chatbot.BotCallbackDataModel
	MsgType        string
	Permission     int
	IsGroup        bool
	CardInstanceId string
	ReceivedMsgStr string
	ConversationID string
	ImageCodeList  []string
	ImageUrlList   []string
}

var (
	messageQueue    chan *Message
	wg              sync.WaitGroup
	dingSupportType []string
)

func DingVarInit() {
	messageQueue = make(chan *Message, 1000) // 设置队列容量
	dingSupportType = []string{"text", "audio", "picture"}
	wg.Add(1)
	go messageConsumer()
}

func messageConsumer() {
	defer wg.Done()
	for msg := range messageQueue {
		// 处理消息的逻辑
		processMessage(msg)
	}
}
func processMessage(msg *Message) {

	if msg.ReceivedMsgStr != "" {

		userID := msg.Data.SenderId
		// 在这里处理你的消息，例如调用API等
		streamScanner, err := DifyClient.CallAPIStreaming(msg.ReceivedMsgStr, userID, msg)
		if err != nil {
			fmt.Println("Error CallAPIStreaming:", err)
			return
		}
		// 发送卡片
		u, err := uuid.NewUUID()
		if err != nil {
			fmt.Println("生成uuid错误")
			return
		}
		cardInstanceId := u.String()
		msg.CardInstanceId = cardInstanceId
		// send interactive card; 发送交互式卡片
		cardData := fmt.Sprintf(consts.MessageCardTemplate, "", "")
		sendOptions := &dingtalkim_1_0.SendRobotInteractiveCardRequestSendOptions{}
		request := &dingtalkim_1_0.SendRobotInteractiveCardRequest{
			CardTemplateId: tea.String("StandardCard"),
			CardBizId:      tea.String(cardInstanceId),
			CardData:       tea.String(cardData),
			RobotCode:      tea.String(clients.DingtalkClient1.ClientID),
			SendOptions:    sendOptions,
			PullStrategy:   tea.Bool(false),
		}
		if msg.Data.ConversationType == "2" {
			// group chat; 群聊
			fmt.Println("钉钉接收群消息:", msg.Data.Text.Content)
			request.SetOpenConversationId(msg.Data.ConversationId)

		} else {
			// ConversationType == "1": private chat; 单聊
			fmt.Println("钉钉接收私聊消息:", msg.Data.Text.Content)
			receiverBytes, err := json.Marshal(map[string]string{"userId": msg.Data.SenderStaffId})
			if err != nil {
				fmt.Println("私聊序列化失败")
				return
			}
			request.SetSingleChatReceiver(string(receiverBytes))
		}
		_, err = clients.DingtalkClient1.SendInteractiveCard(request)
		if err != nil {
			fmt.Println("发送卡片失败")
			return
		}
		// 接收流返回
		var answerBuilder strings.Builder
		cm := selfutils.NewChannelManager()
		defer func() {
			if !cm.IsClosed() {
				cm.CloseChannel()
			}
		}()

		go func(*selfutils.ChannelManager) {
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
							err := UpdateDingTalkCard(content, cardInstanceId)
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

		for streamScanner.Scan() {
			var event StreamingEvent
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

			err = DifyClient.ProcessEvent(userID, event, &answerBuilder, cm)
			if err != nil {
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
		err = UpdateDingTalkCard(answerBuilder.String(), cardInstanceId)
		if err != nil {
			fmt.Println("Error updating DingTalk card:", err)
		}
		// 记录发送日志
		if clients.PermissionControlInit == 1 {
			conversationID, _ := DifyClient.GetSession(userID)
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

// 初始化钉钉机器人
func StartDingRobot() {

	DingVarInit()
	logger.SetLogger(logger.NewStdTestLogger())
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	topic := os.Getenv("Ding_Topic")

	cli := &client.StreamClient{}
	if os.Getenv("Output_Type") == consts.OutputTypeText {
		//纯文本或markdown输出
		cli = client.NewStreamClient(
			client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)),
			client.WithUserAgent(client.NewDingtalkGoSDKUserAgent()),
			client.WithSubscription(utils.SubscriptionTypeKCallback, topic, chatbot.NewDefaultChatBotFrameHandler(OnChatReceiveText).OnEventReceived),
		)
	} else if os.Getenv("Output_Type") == consts.OutputTypeStream {
		clients.DingTalkStreamClientInit()
		// 流式输出
		cli = client.NewStreamClient(
			client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)))
		cli.RegisterChatBotCallbackRouter(OnChatBotStreamingMessageReceived)
	} else if os.Getenv("Output_Type") == consts.OutputTypeMarkDown {
		clients.DingTalkStreamClientInit()
		// 流式输出
		cli = client.NewStreamClient(
			client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)))
		cli.RegisterChatBotCallbackRouter(OnChatReceiveMarkDown)
	}
	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}

	defer cli.Close()

	select {}
}

func OnChatReceiveText(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	if clients.PermissionControlInit == 1 {
		permission, err := clients.PermissionControl.GetUserPermissionLevel(data.SenderId, data.SenderNick)
		if err != nil {
			fmt.Println("OnChatReceive 异常")
			return nil, nil
		}
		fmt.Print(permission)
		if permission == 0 {
			fmt.Println("对不起，没有基础权限，请申请")
		} else if permission == -1 {
			fmt.Println("对不起，没有基础权限，请申请")
		}
	}
	replyMsgStr := strings.TrimSpace(data.Text.Content)
	replier := chatbot.NewChatbotReplier()

	conversationID, exists := DifyClient.GetSession(data.SenderId)
	if exists {
		fmt.Println("Conversation ID for user:", data.SenderId, "is", conversationID)
	} else {
		conversationID = ""
		fmt.Println("No conversation ID found for user:", data.SenderId)
	}

	res, err := DifyClient.CallAPIBlock(replyMsgStr, conversationID, data.SenderId)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(res)

	if err := replier.SimpleReplyText(ctx, data.SessionWebhook, []byte(res)); err != nil {
		return nil, err
	}
	return []byte(""), nil

}

func OnChatBotStreamingMessageReceived(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	// create an uniq card id to identify a card instance while updating
	// see: https://open.dingtalk.com/document/orgapp/robots-send-interactive-cards (cardBizId)
	// 数据过滤
	replier := chatbot.NewChatbotReplier()
	permission := 0
	var err error
	if clients.PermissionControlInit == 1 {
		permission, err = clients.PermissionControl.GetUserPermissionLevel(data.SenderId, data.SenderNick)
		if err != nil {
			fmt.Println("OnChatReceive 异常")
			res := "服务器内部异常"
			if err := replier.SimpleReplyText(ctx, data.SessionWebhook, []byte(res)); err != nil {
				return nil, err
			}
			return nil, err
		}
		fmt.Print(permission)
		if permission == 0 || permission == -1 {
			fmt.Println("对不起，没有基础权限，请申请")
			res := "对不起，没有基础权限，请申请"
			if err := replier.SimpleReplyText(ctx, data.SessionWebhook, []byte(res)); err != nil {
				return nil, err
			}
			return nil, nil
		}

	}

	if !selfutils.StringInSlice(data.Msgtype, dingSupportType) {

		res := "不支持的消息格式"
		if err := replier.SimpleReplyText(ctx, data.SessionWebhook, []byte(res)); err != nil {
			return nil, err
		}
		return nil, nil
	}
	receivedMsgStr := ""
	imageCodeList := []string{}
	imageUrlList := []string{}
	//robotClient := robot_1_0.Client{}
	switch data.Msgtype {
	case consts.ReceivedTypeText:

		receivedMsgStr = strings.TrimSpace(data.Text.Content)
		fmt.Printf("[DingTalk]receive text msg: %s\n", receivedMsgStr)
	case consts.ReceivedTypeVoice:
		fmt.Printf("[DingTalk]receive voice msg: %s\n", data.Content)
		for key, value := range data.Content.(map[string]interface{}) {
			if key == "recognition" {
				recognitionText := value.(string)
				fmt.Println(recognitionText)
				//data.Text.Content = recognitionText
				//if !selfutils.ContainsKeywords(recognitionText, consts.VoicePrefix) {
				//	return []byte(""), nil
				//}
				receivedMsgStr = recognitionText

			}
		}
	//receivedMsgStr
	case consts.ReceivedTypeImage:
		fmt.Printf("[DingTalk]receive image msg: %s", receivedMsgStr)
		for key, value := range data.Content.(map[string]interface{}) {
			if key == "downloadCode" {
				downloadCode := value.(string)
				fmt.Println(downloadCode)
				imageCodeList = append(imageCodeList, downloadCode)
				//RobotMessageFileDownloadWithOptions
				DownloadReq := robot_1_0.RobotMessageFileDownloadRequest{
					DownloadCode: &downloadCode,
				}

				download, err := clients.DingtalkClient1.RobotMessageFileDownload(&DownloadReq)
				if err != nil {
					return nil, err
				}
				fmt.Println(*download.Body.DownloadUrl)
				if download.Body.DownloadUrl != nil {
					imageUrlList = append(imageUrlList, *download.Body.DownloadUrl)
				}
			}
		}

	}
	// 将消息放入队列
	messageQueue <- &Message{
		Ctx:            ctx,
		Data:           data,
		MsgType:        data.Msgtype,
		Permission:     permission,
		ReceivedMsgStr: receivedMsgStr,
		ImageCodeList:  imageCodeList,
		ImageUrlList:   imageUrlList,
	}

	return []byte(""), nil
}

func OnChatReceiveMarkDown(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	if clients.PermissionControlInit == 1 {
		permission, err := clients.PermissionControl.GetUserPermissionLevel(data.SenderId, data.SenderNick)
		if err != nil {
			fmt.Println("OnChatReceive 异常")
			return nil, nil
		}
		fmt.Print(permission)
		if permission == 0 {
			fmt.Println("对不起，没有基础权限，请申请")

		} else if permission == -1 {
			fmt.Println("对不起，没有基础权限，请申请")
		}
	}
	replyMsgStr := strings.TrimSpace(data.Text.Content)
	replier := chatbot.NewChatbotReplier()

	conversationID, exists := DifyClient.GetSession(data.SenderId)
	if exists {
		fmt.Println("Conversation ID for user:", data.SenderId, "is", conversationID)
	} else {
		conversationID = ""
		fmt.Println("No conversation ID found for user:", data.SenderId)
	}

	res, err := DifyClient.CallAPIBlock(replyMsgStr, conversationID, data.SenderId)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(res)
	if err := replier.SimpleReplyMarkdown(ctx, data.SessionWebhook, []byte(""), []byte(res)); err != nil {
		return nil, err
	}

	return []byte(""), nil

}

func UpdateDingTalkCard(content string, cardInstanceId string) error {
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
