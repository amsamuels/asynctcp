package message

import "fmt"

// Queue represents a message queue with multiple named channels.
type Queue struct {
	channels    map[string]map[string]struct{} // Channels with ordered messages.
	chanHandler chan Handler
}

const (
	S  = "SUB"
	P  = "PUB"
	UN = "UNSUB"
)

type Handler struct {
	Action          string
	RemoteIPAddress string // Remote address of the subscriber.
	Topic           string
	Message         string
	Res             chan Response
}

type Response struct {
	Error       error
	Subscribers []string // List of subscriber IDs (for PUB action).
}

func NewQueue() *Queue {
	qu := &Queue{
		channels:    make(map[string]map[string]struct{}, 0),
		chanHandler: make(chan Handler),
	}

	go func() {
		for msg := range qu.chanHandler {
			qu.processMsg(msg)
		}
	}()

	return qu
}

func (qu *Queue) processMsg(msg Handler) {
	switch msg.Action {
	case "SUB":
		if qu.channels[msg.Topic] == nil {
			qu.channels[msg.Topic] = make(map[string]struct{})
		}
		qu.channels[msg.Topic][msg.RemoteIPAddress] = struct{}{}
		msg.Res <- Response{Error: nil}

	case "PUB":
		if qu.channels[msg.Topic] == nil {
			msg.Res <- Response{Error: fmt.Errorf("topic %s doesn't exist", msg.Topic)}
			return
		}

		subscribers := make([]string, 0)
		for id := range qu.channels[msg.Topic] {
			subscribers = append(subscribers, id)
		}
		msg.Res <- Response{Error: nil, Subscribers: subscribers}

	case "UNSUB":
		if qu.channels[msg.Topic] != nil {
			delete(qu.channels[msg.Topic], msg.RemoteIPAddress)
			if len(qu.channels[msg.Topic]) == 0 {
				delete(qu.channels, msg.Topic) // Cleanup empty topics.
			}
			msg.Res <- Response{Error: nil}
			return
		}
		msg.Res <- Response{Error: fmt.Errorf("ID %s is not subscribed to topic %s", msg.RemoteIPAddress, msg.Topic)}
	}
}

// Public methods for the Queue
func (qu *Queue) SUB(clientID, topic string) error {
	msg := Handler{Action: "SUB", RemoteIPAddress: clientID, Topic: topic, Res: make(chan Response)}
	qu.chanHandler <- msg
	resp := <-msg.Res
	return resp.Error
}

func (qu *Queue) PUB(topic, message string) ([]string, error) {
	msg := Handler{Action: "PUB", Topic: topic, Message: message, Res: make(chan Response)}
	qu.chanHandler <- msg
	resp := <-msg.Res
	return resp.Subscribers, resp.Error
}

func (qu *Queue) UNSUB(clientID, topic string) error {
	msg := Handler{Action: "UNSUB", RemoteIPAddress: clientID, Topic: topic, Res: make(chan Response)}
	qu.chanHandler <- msg
	resp := <-msg.Res
	return resp.Error
}
