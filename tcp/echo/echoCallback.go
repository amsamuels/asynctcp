package echo

import (
	"asynctcp/tcp"
	"log"
)

// package allows us to controll how we handle each of the connections
// vet the connections and make sure it is a secure connection

type EchoCallback struct{}

func (ec *EchoCallback) OnConnected(conn *tcp.TCPConn) {
	log.Println("new conn: ", conn.GetRemoteIPAddress())
}
func (ec *EchoCallback) OnMessage(conn *tcp.TCPConn, p tcp.Packet) {
	log.Printf("receive: %s", string(p.Bytes()))
	conn.AsyncWritePacket(p)
}
func (ec *EchoCallback) OnDisconnected(conn *tcp.TCPConn) {
	log.Printf("%s disconnected\n", conn.GetRemoteIPAddress())
}

func (ec *EchoCallback) OnError(err error) {
	log.Println(err)
}
