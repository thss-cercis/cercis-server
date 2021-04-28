package sms

import (
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"math/rand"
	"time"
)

var c *dysmsapi.Client
var signName string
var templateCode string

// Init 初始化 sms 服务
func Init(regionId, accessKeyId, accessKeySecret, _signName, _templateCode string) {
	signName = _signName
	templateCode = _templateCode
	client, err := dysmsapi.NewClientWithAccessKey(regionId, accessKeyId, accessKeySecret)
	if err == nil {
		c = client
	}
}

// GetClient 获得 sms 服务的 client
func GetClient() (client *dysmsapi.Client, ok bool) {
	if c == nil {
		return nil, false
	}
	return c, true
}

func SendSMS(client *dysmsapi.Client, phone string, code string) (response *dysmsapi.SendSmsResponse, err error) {
	if client == nil {
		return nil, errors.New("nil pointer of sms client")
	}
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"

	request.PhoneNumbers = phone
	request.SignName = signName
	request.TemplateCode = templateCode
	request.TemplateParam = fmt.Sprintf("{\"code\":\"%v\"}", code)

	response, err = client.SendSms(request)
	return
}

// NewRandomCode 随机生成一个 sms code
func NewRandomCode() string {
	rand.Seed(time.Now().Unix())
	b := make([]byte, letterLen)
	for i := range b {
		b[i] = letterPool[rand.Intn(len(letterPool))]
	}
	return string(b)
}
