package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Bittorrent message ID
type messageID uint8

const (
	// MsgChoke chokes the receiver
	MsgChoke messageID = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke messageID = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested messageID = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested messageID = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave messageID = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield messageID = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest messageID = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece messageID = 7
	// MsgCancel cancels a request
	MsgCancel messageID = 8
	// MsgExtended identifies following message is using extension protocol
	MsgExtended messageID = 20
)

// Message stores ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
	ExtendedMessage
}

type Bitfield []byte

// Serializes a message for Client to send
func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(len(m.Payload) + 1) // +1 for id
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

// Parses a message from the stream. Returns nil on keep-alive
func Read(r io.Reader) (*Message, error) {
	// Reads the message ID and returns message type

	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil
	}

	// Read the entire message
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	m := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}

	// Check if the message is extended
	if m.ID == MsgExtended {
		m.ExtendedMessage = ExtendedMessage{
			ExtID:   ExtMsgID(m.Payload[0]),
			Payload: m.Payload[1:],
		}
	}

	return &m, nil
}

func FormatRequest(pieceIndex, beginByte, length uint) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(payload[4:8], uint32(beginByte))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

// ParseHave parses a HAVE message
func ParseHave(msg *Message) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length 4, got length %d", len(msg.Payload))
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

func ParsePiece(pieceIndex int, pieceBuffer []byte, msg *Message) (bytesRead int, err error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected PIECE (ID %d), got ID %d", MsgPiece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload too short. %d < 8", len(msg.Payload))
	}

	parsedIndex := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedIndex != pieceIndex {
		return 0, fmt.Errorf("expected pieceIndex %d, got %d", pieceIndex, parsedIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(pieceBuffer) {
		return 0, fmt.Errorf("begin offset too high. %d >= %d", begin, len(pieceBuffer))
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(pieceBuffer) {
		return 0, fmt.Errorf("data too long [%d] for offset %d with length %d", len(data), begin, len(pieceBuffer))
	}

	// Copy the payload of the piece into the buffer
	copy(pieceBuffer[begin:], data)
	return len(data), err
}

// Bitfield with each piece/index the sender has downloaded.
// if bit x is set -> piece index x is downloaded
// Spare bits at the end are set to zero (padding)
// Should be stored with the peer to determine which pieces to request from the peer
func ParseBitfield(msg *Message) Bitfield {
	return msg.Payload
}

func (b Bitfield) HasPiece(pieceIndex int) bool {
	// Find the byte index
	byteIndex := pieceIndex / 8
	bitIndex := pieceIndex % 8

	return b[byteIndex]&(1<<bitIndex) != 0
}

func (m *Message) TypeString() string {
	switch m.ID {
	case MsgChoke:
		return "choke"
	case MsgUnchoke:
		return "unchoke"
	case MsgInterested:
		return "interested"
	case MsgNotInterested:
		return "not interested"
	case MsgHave:
		return "have"
	case MsgBitfield:
		return "bitfield"
	case MsgRequest:
		return "request"
	case MsgPiece:
		return "piece"
	case MsgCancel:
		return "cancel"
	case MsgExtended:

		return "extended"
	default:
		return "unknown"
	}
}
