package clients

import (
	"bytes"
	"ding/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type questionLogClient struct {
	BaseUrl string
}

var QuestionLogCli questionLogClient

func QuestionLogInit() (err error) {
	QuestionLogCli = questionLogClient{
		BaseUrl: os.Getenv("Permission_Service"),
	}
	return nil
}

func (p *questionLogClient) SendQueryRecord(queryLog models.Question) error {
	url := p.BaseUrl + "/questions/add/"
	jsonData, err := json.Marshal(queryLog)
	if err != nil {
		fmt.Printf("Failed to marshal JSON: %v\n", err)
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("query log added successfully")
		return nil
	} else {
		fmt.Printf("Failed to add query log: %s\n", resp.Status)
		var errResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			fmt.Println("Error response:", errResp)
			return err
		}
		return errors.New("异常")
	}
}
