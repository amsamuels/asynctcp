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
	callback := echo.NewEchoCallback(queue)

	srv := tcp.NewAsyncTCPServer(callback, &echo.EchoProtocol{})

	log.Println("start listen...")
	log.Println(srv.Run("localhost:9001"))

	<-c
	srv.Close()

}
