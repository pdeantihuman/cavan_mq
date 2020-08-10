package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/google/logger"
	"github.com/lucas-clemente/quic-go"
	"io"
	"os"
)

const addr = "localhost:4242"
const message = "foobar"

var verbose = flag.Bool("verbose", false, "显示详细信息")

const logPath = "logs/client.log"

func main() {
	flag.Parse()

	lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
	}
	defer lf.Close()

	defer logger.Init("LoggerExample", *verbose, true, lf).Close()
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	session, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		logger.Fatalf("dial失败: %v", err)
	}

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		logger.Fatalf("打开流失败: %v", err)
	}

	fmt.Printf("Client: Sending '%s'\n", message)
	_, err = stream.Write([]byte(message))
	if err != nil {
		logger.Fatalf("写入输出流失败: %v", err)
	}

	buf := make([]byte, len(message))
	_, err = io.ReadFull(stream, buf)
	if err != nil {
		logger.Fatalf("读取输入流失败: %v", err)
	}
	fmt.Printf("Client: Got '%s'\n", buf)
}
