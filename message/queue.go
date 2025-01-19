package message

import "fmt"

// Queue represents a message queue with multiple named channels.
type Queue struct {
	channels    map[string]map[chan string]struct{} // Channels with ordered messages.
	chanHandler chan Handler
}

const (
	S  = "SUB"
	P  = "PUB"
	UN = "UNSUB"
)

type Handler struct {
	Action  string
	SUBS    chan string
	Topic   string
	Message string
	Res     chan error
}

func NewQueue() *Queue {
	qu := &Queue{
		channels:    make(map[string]map[chan string]struct{}, 0),
		chanHandler: make(chan Handler),
	}

	go func() {
		for msg := range qu.chanHandler {
			fmt.Print(msg)
			qu.processMsg(msg)
		}
	}()

	return qu
}

func (qu *Queue) processMsg(msg Handler) {
	switch msg.Action {
	case S:
		if qu.channels[msg.Topic] != nil {
			qu.channels[msg.Topic][msg.SUBS] = struct{}{}
			msg.Res <- nil
			return
		}
	case P:
		if qu.channels[msg.Topic] == nil {
			msg.Res <- fmt.Errorf("chan doesn't exist ")
			return
		}

		for ch := range qu.channels[msg.Topic] {
			ch <- msg.Message
		}
	case UN:
		if qu.channels[msg.Topic] != nil {
			delete(qu.channels[msg.Topic], msg.SUBS)
			return
		}
		msg.Res <- fmt.Errorf("not subbed to topic")
	}

}

func (qu *Queue) SUB(clientID chan string, channel string) error {
	var msg = Handler{Action: S, SUBS: clientID, Topic: channel}
	qu.chanHandler <- msg
	return <-msg.Res
}

func (qu *Queue) PUB(channel string, message string) error {
	var msg = Handler{Action: P, Topic: channel, Message: message}
	qu.chanHandler <- msg
	return <-msg.Res
}

func (qu *Queue) UNSUB(clientID chan string, channel string) error {
	var msg = Handler{Action: UN, SUBS: clientID, Topic: channel}
	qu.chanHandler <- msg
	return <-msg.Res
}
