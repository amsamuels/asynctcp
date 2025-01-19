package main

import (
	"asynctcp/tcp"
	"bufio"
	"io"
	"log"
	"time"
)

func main() {
	srv := tcp.NewAsyncTCPServer("localhost:9001", &EchoCallback{}, &EchoProtocol{})
	srv.SetReadDeadline(time.Second * 10)
	log.Println("start listen...")
	log.Println(srv.ListenAndServe())
	select {}
}

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

type EchoProtocol struct {
	data []byte
}

func (ep *EchoProtocol) Bytes() []byte {
	return ep.data
}

func (ep *EchoProtocol) ReadPacket(reader io.Reader) (tcp.Packet, error) {
	rd := bufio.NewReader(reader)
	bytesData, err := rd.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return &EchoProtocol{data: bytesData}, nil
}
func (ep *EchoProtocol) WritePacket(writer io.Writer, msg tcp.Packet) error {
	_, err := writer.Write(msg.Bytes())
	return err
}
