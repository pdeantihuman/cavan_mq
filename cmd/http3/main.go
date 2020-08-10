package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"github.com/google/logger"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
)

var logPath = flag.String("log-path", "logs/http3.log", "http3的日志文件路径")

var verbose = flag.Bool("verbose", false, "输出详细")

func setupHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("/echo 读取body失败: %v", err)
		}
		_, err = w.Write(body)
		if err != nil {
			logger.Errorf("/echo 写入输出流失败: %v", err)
		}
	})
	return mux
}

func main() {
	flag.Parse()
	lf, err := os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()

	defer logger.Init("LoggerExample", *verbose, true, lf).Close()
	handler := setupHandler()
	quicConf := &quic.Config{}
	done := make(chan int)
	go func() {
		var err error
		server := http3.Server{
			Server:     &http.Server{Handler: handler, Addr: "127.0.0.1:1252"},
			QuicConfig: quicConf,
		}
		err = server.ListenAndServe()
		if err != nil {
			logger.Errorf("http3服务器启动失败: %v", err)
		}
		close(done)
	}()
	<-done
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() (*tls.Config, error) {
	// 从随机数生成一对 RSA 密钥
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		logger.Errorf("RSA密钥生成失败: %v", err)
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	// 将 rsa 私钥转换为 PKCS#1 格式
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}, nil
}
