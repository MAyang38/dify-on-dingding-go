package clients

import (
	"bytes"
	"ding/models"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// 定义 ResponseData 结构体来匹配 JSON 数据
type PermissionResponse struct {
	TotalCount int           `json:"totalCount"`
	Users      []models.User `json:"users"`
}
type permissionControl struct {
	BaseUrl string
}

var PermissionControl permissionControl
var PermissionControlInit = 0

func PermissionInit() (err error) {
	PermissionControl = permissionControl{
		BaseUrl: os.Getenv("Permission_Service"),
	}
	PermissionControlInit, err = strconv.Atoi(os.Getenv("Permission_Control_Init"))
	if err != nil {
		return err
	}
	return nil
}

func (p *permissionControl) GetUserPermissionLevel(userID string) (int, error) {

	encodedUserID := url.QueryEscape(userID)
	url := fmt.Sprintf("%s/users?user_id=%s", p.BaseUrl, encodedUserID)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Request failed with status: %s\n", resp.Status)
		return -1, errors.New("Request failed with status")
	}
	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
		return -1, err
	}

	// 解析 JSON 数据
	var data PermissionResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
		return -1, err
	}
	if len(data.Users) > 0 {
		return data.Users[0].PermissionLevel, nil
	}

	return p.AddUserPermissionLevel(userID)

}

func (p *permissionControl) AddUserPermissionLevel(userID string) (int, error) {
	url := p.BaseUrl + "/users"
	defaultPermissionLevel := 1001 // 替换为你的配置获取逻辑

	user := models.User{
		UserID:          userID,
		PermissionLevel: defaultPermissionLevel,
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		return -1, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("User added successfully")
		return defaultPermissionLevel, nil
	} else {
		fmt.Printf("Failed to add user: %s\n", resp.Status)
		var errResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			fmt.Println("Error response:", errResp)
			return -1, err
		}
		return -1, errors.New("异常")
	}
}
