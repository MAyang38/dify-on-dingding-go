package clients

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkim_1_0 "github.com/alibabacloud-go/dingtalk/im_1_0"
	dingtalkoauth2_1_0 "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	"github.com/alibabacloud-go/dingtalk/robot_1_0"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"os"
	"time"
)

type DingTalkClient struct {
	ClientID      string
	clientSecret  string
	accessToken   string
	tokenExpireAt time.Time
	imClient      *dingtalkim_1_0.Client
	oauthClient   *dingtalkoauth2_1_0.Client
	robotClient   *robot_1_0.Client
}

var (
	DingtalkClient1 *DingTalkClient = nil
)

func DingTalkStreamClientInit() {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	DingtalkClient1 = NewDingTalkClient(clientId, clientSecret)
}
func NewDingTalkClient(clientId, clientSecret string) *DingTalkClient {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	imClient, _ := dingtalkim_1_0.NewClient(config)
	oauthClient, _ := dingtalkoauth2_1_0.NewClient(config)
	robotClient, _ := robot_1_0.NewClient(config)
	return &DingTalkClient{
		ClientID:     clientId,
		clientSecret: clientSecret,
		imClient:     imClient,
		oauthClient:  oauthClient,
		robotClient:  robotClient,
	}
}

func (c *DingTalkClient) GetAccessToken() (string, error) {
	// 检查当前 token 是否过期
	if time.Now().Before(c.tokenExpireAt) {
		return c.accessToken, nil
	}
	request := &dingtalkoauth2_1_0.GetAccessTokenRequest{
		AppKey:    tea.String(c.ClientID),
		AppSecret: tea.String(c.clientSecret),
	}
	response, tryErr := func() (_resp *dingtalkoauth2_1_0.GetAccessTokenResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_resp, _err := c.oauthClient.GetAccessToken(request)
		if _err != nil {
			return nil, _err
		}

		return _resp, nil
	}()
	if tryErr != nil {
		return "", tryErr
	}
	c.accessToken = *response.Body.AccessToken
	c.tokenExpireAt = time.Now().Add(1 * time.Hour)

	return *response.Body.AccessToken, nil
}

func (c *DingTalkClient) SendInteractiveCard(request *dingtalkim_1_0.SendRobotInteractiveCardRequest) (*dingtalkim_1_0.SendRobotInteractiveCardResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	headers := &dingtalkim_1_0.SendRobotInteractiveCardHeaders{
		XAcsDingtalkAccessToken: tea.String(accessToken),
	}
	response, tryErr := func() (_resp *dingtalkim_1_0.SendRobotInteractiveCardResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_resp, _e = c.imClient.SendRobotInteractiveCardWithOptions(request, headers, &util.RuntimeOptions{})
		if _e != nil {
			return
		}
		return
	}()
	if tryErr != nil {
		return nil, tryErr
	}
	return response, nil
}

func (c *DingTalkClient) UpdateInteractiveCard(request *dingtalkim_1_0.UpdateRobotInteractiveCardRequest) (*dingtalkim_1_0.UpdateRobotInteractiveCardResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	headers := &dingtalkim_1_0.UpdateRobotInteractiveCardHeaders{
		XAcsDingtalkAccessToken: tea.String(accessToken),
	}
	response, tryErr := func() (_resp *dingtalkim_1_0.UpdateRobotInteractiveCardResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_resp, _e = c.imClient.UpdateRobotInteractiveCardWithOptions(request, headers, &util.RuntimeOptions{})
		if _e != nil {
			return
		}
		return
	}()
	if tryErr != nil {
		return nil, tryErr
	}
	return response, nil
}

func (c *DingTalkClient) RobotMessageFileDownload(request *robot_1_0.RobotMessageFileDownloadRequest) (*robot_1_0.RobotMessageFileDownloadResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	headers := &robot_1_0.RobotMessageFileDownloadHeaders{
		XAcsDingtalkAccessToken: tea.String(accessToken),
	}
	request.RobotCode = &c.ClientID
	response, tryErr := func() (_resp *robot_1_0.RobotMessageFileDownloadResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		_resp, _e = c.robotClient.RobotMessageFileDownloadWithOptions(request, headers, &util.RuntimeOptions{})

		//_resp, _e = c.imClient.UpdateRobotInteractiveCardWithOptions(request, headers, &util.RuntimeOptions{})
		if _e != nil {
			return
		}
		return
	}()
	if tryErr != nil {
		return nil, tryErr
	}
	return response, nil
}
