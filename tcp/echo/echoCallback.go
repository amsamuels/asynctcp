package echo

import (
	"asynctcp/message"
	"asynctcp/tcp"
	"log"
	"strings"
)

// package allows us to controll how we handle each of the connections
// vet the connections and make sure it is a secure connection

type EchoCallback struct {
	queue *message.Queue // Message queue instance.
}

func NewEchoCallback(queue *message.Queue) *EchoCallback {
	return &EchoCallback{queue: queue}
}

func (ec *EchoCallback) OnConnected(conn *tcp.TCPConn) {
	//conn.setReadDeadline(time.Second * 300)
	log.Println("new conn: ", conn.GetRemoteIPAddress())
}
func (ec *EchoCallback) OnMessage(conn *tcp.TCPConn, p tcp.Packet) {
	message := string(p.Bytes()) // Convert the packet to a string.

	// Parse the message into command and arguments.
	parts := strings.SplitN(message, " ", 2)
	if len(parts) < 2 {
		ec.respondWithError(conn, "Invalid message format. Expected: <command> <data>\n")
		return
	}

	command, data := parts[0], parts[1]
	switch command {
	case "S":
		log.Printf("Subscribing to topic: %s", data)
		ec.queue.SUB(conn.GetRemoteIPAddress(), data)
		ec.respondWithSuccess(conn, "Subscribed to topic: "+data+"\n")
	case "P":
		topicAndMessage := strings.SplitN(data, " ", 2)
		if len(topicAndMessage) < 2 {
			log.Println("Invalid publish format")
			return
		}
		topic, message := topicAndMessage[0], topicAndMessage[1]
		log.Printf("Publishing to topic: %s, message: %s", topic, message)
		ec.queue.PUB(topic, message)
	case "UNS":
		ec.queue.UNSUB(conn.GetRemoteIPAddress(), data)
		ec.respondWithSuccess(conn, "Unsubscribed from topic: "+data+"\n")
	default:
		ec.respondWithError(conn, "Unknown command: "+command+"\n")
	}
}
func (ec *EchoCallback) OnDisconnected(conn *tcp.TCPConn) {
	log.Printf("%s disconnected\n", conn.GetRemoteIPAddress())
}

func (ec *EchoCallback) OnError(err error) {
	log.Println(err)
}

func (ec *EchoCallback) respondWithError(conn *tcp.TCPConn, errorMessage string) {
	errPacket := tcp.NewDefaultPacket(tcp.TypeMessage, []byte(errorMessage))
	if err := conn.AsyncWritePacket(errPacket); err != nil {
		log.Printf("Failed to send error response to client: %v", err)
	}
}

func (ec *EchoCallback) respondWithSuccess(conn *tcp.TCPConn, successMessage string) {
	pkt := tcp.NewDefaultPacket(tcp.TypeMessage, []byte(successMessage))
	if err := conn.AsyncWritePacket(pkt); err != nil {
		log.Printf("Failed to send success response to client: %v", err)
	}
}
