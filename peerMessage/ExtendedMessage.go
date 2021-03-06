package message

import (
	"bytes"
	"fmt"
	"math"

	"github.com/jackpal/bencode-go"
)

type ExtMsgID uint8

type MetadataMsgType uint8

const (
	ExtHandshake ExtMsgID        = 0
	MetaRequest  MetadataMsgType = 0
	MetaData     MetadataMsgType = 1
	MetaReject   MetadataMsgType = 2
)

type Map map[string]int

type ExtendedMessage struct {
	ExtID   ExtMsgID
	Payload []byte
}

const METADATA_PAYLOAD_SIZE = 1024 * 16

type MetaReq struct {
	MsgType int `bencode:"msg_type"`
	Piece   int `bencode:"piece"`
}
type MetaDataData struct {
	MsgType int    `bencode:"msg_type"`
	Piece   int    `bencode:"piece"`
	Size    int    `bencode:"total_size"`
	Payload []byte `bencode:"-"` // This is the payload, but it not bencode encoded, so should be ignored when marshalling/unmarshalling
}

/* Headers:
size			description
uint8_t		extended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
*/

func (m *ExtendedMessage) Serialize() []byte {
	buf := make([]byte, len(m.Payload)+1)
	// 1st byte is the extended message ID
	buf[0] = byte(m.ExtID)
	copy(buf[1:], m.Payload)
	msg := Message{
		ID:      MsgExtended,
		Payload: buf,
	}

	return msg.Serialize()
}

// The metadata extension uses the extension protocol (specified in BEP 0010 ) to advertize its existence.
// It adds the "ut_metadata" entry to the "m" dictionary in the extension header hand-shake message.
// This identifies the message code used for this message.
// It also adds "metadata_size" to the handshake message (not the "m" dictionary)
// specifying an integer value of the number of bytes of the metadata.

func ParseMetadata(msg ExtendedMessage, extensions Map) (MetaDataData, error) {
	var piece MetaDataData
	if int(msg.ExtID) != extensions["ut_metadata"] {
		return piece, fmt.Errorf("Bad extended message ID. Expected %v got: %v\n", msg.ExtID, extensions["ut_metadata"])
	}

	// 1. Parse the metadata response
	r := bytes.NewReader(msg.Payload[:])
	if err := bencode.Unmarshal(r, &piece); err != nil {
		return piece, err
	}

	// 2. Check the message type
	if piece.MsgType != int(MetaData) {
		return piece, fmt.Errorf("Bad message type. Expected %v got: %v\n", MetaData, piece.MsgType)
	}

	// 3. calculate total metadata pieces
	totalPieces := int(math.Ceil(float64(piece.Size) / float64(METADATA_PAYLOAD_SIZE)))

	// 4. Calculate the payload offset (room for improvement??)
	var payloadOffset int
	if piece.Piece < totalPieces-1 {
		payloadOffset = len(msg.Payload) - METADATA_PAYLOAD_SIZE
	} else {
		payloadOffset = len(msg.Payload) - (piece.Size % METADATA_PAYLOAD_SIZE)
	}

	// 5. Read the rest of the msg payload & copy into piece payload
	piece.Payload = make([]byte, METADATA_PAYLOAD_SIZE) // 16KB
	copy(piece.Payload[:], msg.Payload[payloadOffset:])

	return piece, nil
}
