package main

import (
	"asynctcp/message"
	"asynctcp/tcp"
	"asynctcp/tcp/echo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	queue := message.NewQueue()
	connBucket := tcp.NewTCPConnBucket()
	callback := echo.NewEchoCallback(queue, connBucket)
	var proto echo.EchoProtocol
	srv := tcp.NewAsyncTCPServer(callback, &proto)

	log.Println("start listen...")
	log.Println(srv.Run("localhost:9001"))

	<-c
	srv.Close()

}
