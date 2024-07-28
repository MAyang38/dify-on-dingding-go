package audio

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	lfasrHost    = "https://raasr.xfyun.cn/v2/api"
	apiUpload    = "/upload"
	apiGetResult = "/getResult"
)

type RequestApi struct {
	AppID          string
	SecretKey      string
	UploadFilePath string
	Timestamp      string
	Signa          string
}

// SpeechResult 代表整个识别结果的结构
type SpeechResult struct {
	Code     string        `json:"code"`
	Content  SpeechContent `json:"content"`
	DescInfo string        `json:"descInfo"`
}

// SpeechContent 代表 result 部分的结构
type SpeechContent struct {
	OrderInfo   OrderInfo `json:"orderInfo"`
	OrderResult string    `json:"orderResult"`
}

// OrderInfo 代表 orderInfo 部分的结构
type OrderInfo struct {
	ExpireTime       float64 `json:"expireTime"`
	FailType         int     `json:"failType"`
	OrderID          string  `json:"orderId"`
	OriginalDuration int     `json:"originalDuration"`
	RealDuration     int     `json:"realDuration"`
	Status           int     `json:"status"`
}

// SpeechResultData 代表 orderResult 中的数据结构
type SpeechResultData struct {
	Lattice  []LatticeItem  `json:"lattice"`
	Lattice2 []Lattice2Item `json:"lattice2"`
}

// LatticeItem 代表 lattice 部分的结构
type LatticeItem struct {
	Json1Best string `json:"json_1best"`
}

// Lattice2Item 代表 lattice2 部分的结构
type Lattice2Item struct {
	Lid       string       `json:"lid"`
	End       string       `json:"end"`
	Begin     string       `json:"begin"`
	Json1Best Lattice2Best `json:"json_1best"`
}

// Lattice2Best 代表 lattice2 中 json_1best 部分的结构
type Lattice2Best struct {
	St SpeechText `json:"st"`
}

// SpeechText 代表语音识别的详细文本结构
type SpeechText struct {
	Sc string     `json:"sc"`
	Pa string     `json:"pa"`
	Rt []SpeechRT `json:"rt"`
}

// SpeechRT 代表 rt 部分的结构
type SpeechRT struct {
	Ws []SpeechWS `json:"ws"`
}

// SpeechWS 代表 ws 部分的结构
type SpeechWS struct {
	Cw []SpeechCW `json:"cw"`
}

// SpeechCW 代表 cw 部分的结构
type SpeechCW struct {
	W  string `json:"w"`
	Wp string `json:"wp"`
	Wc string `json:"wc"`
}

func extractTextFromResult(orderResult string) (string, error) {
	var result SpeechResultData

	err := json.Unmarshal([]byte(orderResult), &result)
	if err != nil {
		return "", err
	}

	var text string

	// 从 lattice2 中提取文本
	for _, lattice2 := range result.Lattice2 {
		st := lattice2.Json1Best.St
		for _, rtItem := range st.Rt {
			for _, wsItem := range rtItem.Ws {
				for _, cwItem := range wsItem.Cw {
					text += cwItem.W
				}
			}
		}
	}

	return text, nil
}

func (api *RequestApi) getSigna() string {
	appid := api.AppID
	secretKey := api.SecretKey
	ts := api.Timestamp

	// 1. 获取 baseString

	baseString := appid + ts

	// 2. 对 baseString 进行 MD5
	md5Hash := md5.Sum([]byte(baseString))
	md5HashString := fmt.Sprintf("%x", md5Hash)

	// 3. 使用 secretKey 对 MD5 结果进行 HMAC-SHA1 加密
	hmacHash := hmac.New(sha1.New, []byte(secretKey))
	hmacHash.Write([]byte(md5HashString))
	signa := base64.StdEncoding.EncodeToString(hmacHash.Sum(nil))

	return signa
}

func (api *RequestApi) upload() (map[string]interface{}, error) {
	fmt.Println("上传部分：")
	filePath := api.UploadFilePath
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileLen := fileInfo.Size()
	fileName := fileInfo.Name()

	param := url.Values{}
	param.Add("appId", api.AppID)
	param.Add("signa", api.Signa)
	param.Add("ts", api.Timestamp)
	param.Add("fileSize", strconv.FormatInt(fileLen, 10))
	param.Add("fileName", fileName)
	param.Add("duration", "200")

	url := lfasrHost + apiUpload + "?" + param.Encode()
	fmt.Println("upload参数：", param)
	fmt.Println("upload_url:", url)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	fmt.Println("upload resp:", result)
	return result, nil
}

func (api *RequestApi) getResult() (*SpeechResult, error) {
	uploadResp, err := api.upload()
	if err != nil {
		return nil, err
	}

	orderId := uploadResp["content"].(map[string]interface{})["orderId"].(string)

	param := url.Values{}
	param.Add("appId", api.AppID)
	param.Add("signa", api.Signa)
	param.Add("ts", api.Timestamp)
	param.Add("orderId", orderId)
	param.Add("resultType", "transfer,predict")

	url := lfasrHost + apiGetResult + "?" + param.Encode()
	fmt.Println("查询部分：")
	fmt.Println("get result参数：", param)

	//var result map[string]interface{}
	var result SpeechResult
	status := 3

	for status == 3 {
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}

		fmt.Println(result)
		status = result.Content.OrderInfo.Status
		if status == 4 {
			break
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Println("get_result resp:", result)
	return &result, nil
}

func XunfeiHandler() {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	api := RequestApi{
		AppID:          os.Getenv("XUNFEI_APPID"),
		SecretKey:      os.Getenv("XUNFEI_SecretKey"),
		UploadFilePath: "audio/aigei_com.wav",
		Timestamp:      ts,
		Signa:          "",
	}
	api.Signa = api.getSigna()

	maps, err := api.getResult()
	if err != nil {
		fmt.Println("Error:", err)
	}
	str := maps.Content.OrderResult
	text, err := extractTextFromResult(str)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Extracted Text:", text)
}
