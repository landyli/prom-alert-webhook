package adapter

import (
	"bytes"
	"prom-alert-webhook/alertMessage"
	"prom-alert-webhook/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type dingtalk struct {
	secret      string
	accessToken string
	sendData    string
}

type dingMsg struct {
	MsgType string            `json:"msgtype"`
	Text    map[string]string `json:"text"`
}

const dingWebHook = "https://oapi.dingtalk.com/robot/send?access_token"

func InitDingTalk(secret, accessToken string) *dingtalk {
	return &dingtalk{
		secret:      secret,
		accessToken: accessToken,
	}
}

func (d dingtalk) Cmd(sendData alertMessage.AlertMessage) {
	for _, data := range sendData.Alerts {
		d.sendData = utils.FormatData(data)
		content := formatData(d.sendData)
		var msg = dingMsg{
			MsgType: "text",
			Text:    map[string]string{"content": content},
		}
		d.sendMsg(msg)
	}
}

func (d dingtalk) hmacSha256(stringToSign string) string {
	h := hmac.New(sha256.New, []byte(d.secret))
	h.Write([]byte(stringToSign))
	data := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(data)
}

// 加签
func (d dingtalk) sign() string {
	timestamp := time.Now().UnixNano() / 1e6
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, d.secret)
	sign := d.hmacSha256(stringToSign)
	fmt.Println(sign)
	url := fmt.Sprintf("%s=%s&timestamp=%d&sign=%s", dingWebHook, d.accessToken, timestamp, sign)
	fmt.Println(url)
	return url
}

func (d dingtalk) sendMsg(msg dingMsg) {
	query := url.Values{}
	query.Set("access_token", d.accessToken)
	hookUrl, _ := url.Parse(dingWebHook)
	hookUrl.RawQuery = query.Encode()
	msgContent, _ := json.Marshal(msg)
	//创建一个请求
	//req, err := http.NewRequest("POST", hookUrl.String(), bytes.NewReader(msgContent))
	req, err := http.NewRequest("POST", d.sign(), bytes.NewReader(msgContent))
	if err != nil {
		// handle error
		log.Println("创建新请求失败失败:", err)
	}

	client := &http.Client{}
	//设置请求头
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	//发送请求
	resp, err := client.Do(req)
	//关闭请求
	defer resp.Body.Close()
	if err != nil {
		// handle error
		log.Println("设置请求头content-type失败:", err)
	}
	if resp.StatusCode == http.StatusOK{
		log.Println("发送报警信息到钉钉成功！！！")
	}
}
