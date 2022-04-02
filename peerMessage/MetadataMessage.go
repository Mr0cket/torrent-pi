package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MetadataMessageID uint8

const (
	RequestMetadata MetadataMessageID = 0
)

type ExtendedMessage struct {
	ID      int
	Payload []byte
}

/* Headers:
size			description
uint32_t	length prefix. Specifies the number of bytes for the entire message. (big-endian)
uint8_t		bittorrent message ID, = 20
uint8_t		extended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
*/
func (m *ExtendedMessage) Serialize() []byte {
	length := uint32(len(m.Payload)) // 1 for id, 1 extended message id, 4 for length
	fmt.Println("Serializing message: ", m.ID, " with length: ", length)
	fmt.Println("Payload: ", m.Payload)
	buf := make([]byte, length)
	// binary.BigEndian.PutUint32(buf[:], length)
	// buf[4] = byte(MsgExtended)
	// buf[5] = byte(m.ID)
	copy(buf[:], m.Payload)
	fmt.Println("Serialized message: ", buf)
	return buf
}

func ReadExtension(r io.Reader) (*ExtendedMessage, error) {
	// Reads the message ID and returns message type
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		fmt.Println("Error reading length: ", err)
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil
	}

	// Read the
	messageBuf := make([]byte, length)
	fmt.Println("packet length:", length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	if messageBuf[4] != byte(MsgExtended) {
		return nil, fmt.Errorf("Error: Expected extended message ID 20")

	}

	m := ExtendedMessage{
		ID:      int(messageBuf[5]),
		Payload: messageBuf[6:],
	}
	return &m, nil
}

// The metadata extension uses the extension protocol (specified in BEP 0010 ) to advertize its existence.
// It adds the "ut_metadata" entry to the "m" dictionary in the extension header hand-shake message.
// This identifies the message code used for this message.
// It also adds "metadata_size" to the handshake message (not the "m" dictionary)
// specifying an integer value of the number of bytes of the metadata.
