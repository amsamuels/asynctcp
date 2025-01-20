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
	queue      *message.Queue // Message queue instance.
	connBucket *tcp.TCPConnBucket
}

func NewEchoCallback(queue *message.Queue, connBucket *tcp.TCPConnBucket) *EchoCallback {
	return &EchoCallback{
		queue:      queue,
		connBucket: connBucket,
	}
}

func (ec *EchoCallback) OnConnected(conn *tcp.TCPConn) {
	log.Println("new conn: ", conn.GetRemoteIPAddress())
	ec.connBucket.Put(conn.GetRemoteIPAddress(), conn)
}

func (ec *EchoCallback) OnDisconnected(conn *tcp.TCPConn) {
	log.Printf("%s disconnected\n", conn.GetRemoteIPAddress())
	ec.connBucket.Delete(conn.GetRemoteIPAddress())
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
		Subscribers, _ := ec.queue.PUB(topic)
		ec.respondToAllSubs(Subscribers, ""+message+"\n")
	case "UNS":
		ec.queue.UNSUB(conn.GetRemoteIPAddress(), data)
		ec.respondWithSuccess(conn, "Unsubscribed from topic: "+data+"\n")
	default:
		ec.respondWithError(conn, "Unknown command: "+command+"\n")
	}
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

func (ec *EchoCallback) respondToAllSubs(addresses []string, message string) {
	pkt := tcp.NewDefaultPacket(tcp.TypeMessage, []byte(message))

	for _, address := range addresses {
		conn := ec.connBucket.GetByAddress(address)
		if conn == nil {
			log.Printf("No connection found for address: %s", address)
			continue
		}

		if err := conn.AsyncWritePacket(pkt); err != nil {
			log.Printf("Failed to send message to address %s: %v", address, err)
		}
	}
}
