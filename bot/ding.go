package bot

import (
	"context"
	"ding/clients"
	"ding/consts"
	"encoding/json"
	"fmt"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/utils"
	"os"
	"strings"
)

// 初始化钉钉机器人
func StartDingRobot() {
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
		cli.RegisterChatBotCallbackRouter(OnChatBotStreamingMessageReceived)
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
		permission, err := clients.PermissionControl.GetUserPermissionLevel(data.SenderId)
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
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	cardInstanceId := u.String()

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
	if data.ConversationType == "2" {
		// group chat; 群聊
		fmt.Println("钉钉接收群消息:", data.Text.Content)
		request.SetOpenConversationId(data.ConversationId)
	} else {
		// ConversationType == "1": private chat; 单聊
		fmt.Println("钉钉接收私聊消息:", data.Text.Content)
		receiverBytes, err := json.Marshal(map[string]string{"userId": data.SenderStaffId})
		if err != nil {
			return nil, err
		}
		request.SetSingleChatReceiver(string(receiverBytes))
	}
	_, err = clients.DingtalkClient1.SendInteractiveCard(request)
	if err != nil {
		return nil, err
	}
	receivedMsgStr := strings.TrimSpace(data.Text.Content)
	conversationID, exists := DifyClient.GetSession(data.SenderId)
	if exists {
		fmt.Println("Conversation ID for user:", data.SenderId, "is", conversationID)
	} else {
		conversationID = ""
		fmt.Println("No conversation ID found for user:", data.SenderId)
	}
	err = DifyClient.CallAPIStreaming(receivedMsgStr, conversationID, data.SenderId, cardInstanceId)
	if err != nil {
		return nil, err
	}

	return []byte(""), nil
}

func OnChatReceiveMarkDown(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	if clients.PermissionControlInit == 1 {
		permission, err := clients.PermissionControl.GetUserPermissionLevel(data.SenderId)
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
	if err := replier.SimpleReplyMarkdown(ctx, data.SessionWebhook, []byte("stream-tutorial-go"), []byte(res)); err != nil {
		return nil, err
	}

	return []byte(""), nil

}
