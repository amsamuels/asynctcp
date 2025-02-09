package tcp

import (
	"bytes"
	"encoding/binary"
)

type Packet interface {
	Bytes() []byte
}

type PacketType byte

const (
	TypeCommand PacketType = iota // Represents a command packet.
	TypeMessage                   // Represents a message packet.
	TypeError                     // Represents an error packet.
)

type DefaultPacket struct {
	Type PacketType
	Body []byte
}

func NewDefaultPacket(t PacketType, body []byte) *DefaultPacket {
	return &DefaultPacket{
		Type: t,
		Body: body,
	}
}

func (m *DefaultPacket) Bytes() []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, m.Type)
	binary.Write(&buf, binary.BigEndian, m.Body)
	return buf.Bytes()
}
