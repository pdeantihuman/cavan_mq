package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"github.com/google/logger"
	"github.com/lucas-clemente/quic-go"
	"io"
	"math/big"
	"os"
)

var verbose = flag.Bool("verbose", false, "显示详细信息")

const logPath = "server.log"

const addr = "localhost:4242"

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}
}

// A wrapper for io.Writer that also logs the message.
type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	logger.Infof("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}

func main() {
	flag.Parse()
	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("日志文件打开失败: %v", err)
	}
	defer lf.Close()

	defer logger.Init("LoggerExample", *verbose, true, lf).Close()
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		logger.Fatalf("quic监听发生了错误: %v", err)
	}
	sess, err := listener.Accept(context.Background())
	if err != nil {
		logger.Fatalf("quic监听发生了错误: %v", err)
	}
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		logger.Fatalf("quic监听发生了错误: %v", err)
	}
	// Echo through the loggingWriter
	_, err = io.Copy(loggingWriter{stream}, stream)
	if err != nil {
		logger.Fatalf("quic监听发生了错误: %v", err)
	}
}
