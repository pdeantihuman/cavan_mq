package main

import (
	"crypto/tls"
	"fmt"
	"github.com/google/logger"
	"github.com/lucas-clemente/quic-go/http3"
	"net/http"
)

func main() {
	roundTripper := &http3.RoundTripper{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	}}
	hclient := &http.Client{
		Transport: roundTripper,
	}
	var in string
	for {
		_, err := fmt.Scanln(&in)
		if err != nil {
			logger.Errorf("获取用户输入失败: %v", err)
			continue
		}
		get, err := hclient.Get("https://127.0.0.1:1252/echo")
		if err != nil {
			logger.Errorf("发送get消息失败: %v", err)
			continue
		}
		logger.Infof("接收到消息: %v", get)
	}
}
