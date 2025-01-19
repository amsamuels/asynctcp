package main

import (
	"asynctcp/tcp"
	"asynctcp/tcp/echo"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	srv := tcp.NewAsyncTCPServer(&echo.EchoCallback{}, &echo.EchoProtocol{})
	srv.SetReadDeadline(time.Second * 10)

	log.Println("start listen...")
	log.Println(srv.Run("localhost:9001"))

	<-c
	srv.Close()

}
