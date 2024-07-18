package audio

import (
	"ding/models"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type BaiduVoice struct {
	ClientID     string
	ClientSecret string
	Token        string
	Expire       time.Time
}

type VoiceData struct {
	Format  string      `json:"format"`
	Rate    int         `json:"rate"`
	Channel int         `json:"channel"`
	Token   interface{} `json:"token"`
	DevPid  int         `json:"dev_pid"`
	Cuid    string      `json:"cuid"`
	Len     int         `json:"len"`
	Speech  string      `json:"speech"`
}

var BaiduVoicdeCli *BaiduVoice

const tokenUrlTemplpate = "https://aip.baidubce.com/oauth/2.0/token?client_id=%s&client_secret=%s&grant_type=client_credentials"

func BaiduVoiceInit() {
	BaiduVoicdeCli = &BaiduVoice{
		ClientID:     os.Getenv("BaiduClientId"),
		ClientSecret: os.Getenv("BaiduClientSecret"),
		Expire:       time.Now(),
	}
}
func (c *BaiduVoice) VoiceToText(filePath string) (string, error) {
	url := "https://vop.baidu.com/server_api"
	// 指定本地文件路径
	// 读取文件内容
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// 计算文件大小（字节数）
	fileSize := len(fileBytes)
	fmt.Printf("File size: %d bytes\n", fileSize)

	// 将文件内容转换为Base64编码
	token, err := c.GetAccessToken()
	if err != nil {
		return "", err
	}

	base64String := base64.StdEncoding.EncodeToString(fileBytes)
	voiceData := VoiceData{
		Format:  "pcm",
		Rate:    16000,
		Channel: 1,
		DevPid:  1537,
		Token:   token,
		Cuid:    "TZu3ZUWS8wQQ7Sa3gdYQI4aFGb08xTG1",
		Len:     fileSize,
		Speech:  base64String,
	}
	// 将VoiceData结构体编码为JSON
	jsonData, err := json.Marshal(voiceData)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}
	payload := strings.NewReader(string(jsonData))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	voiceRespBody := models.VoicdeRespBody{}
	_ = json.Unmarshal(body, &voiceRespBody)
	fmt.Println(voiceRespBody)
	if voiceRespBody.ErrNo != 0 {
		return "", errors.New(voiceRespBody.ErrMsg)
	}
	return voiceRespBody.Result[0], nil
}

/**
 * 使用 AK，SK 生成鉴权签名（Access Token）
 * @return string 鉴权签名信息（Access Token）
 */
func (c *BaiduVoice) GetAccessToken() (string, error) {
	if time.Now().Before(c.Expire) {
		return c.Token, nil
	}
	url := "https://aip.baidubce.com/oauth/2.0/token"
	postData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", c.ClientID, c.ClientSecret)
	fmt.Printf(postData)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(body))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	accessTokenObj := models.AccessTokenResponse{}
	_ = json.Unmarshal(body, &accessTokenObj)
	c.Token = accessTokenObj.AccessToken
	c.Expire = time.Now().Add(time.Duration(accessTokenObj.ExpiresIn) * time.Second)
	return accessTokenObj.AccessToken, nil
}
