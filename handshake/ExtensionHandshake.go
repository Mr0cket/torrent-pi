package handshake

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"torrent-pi/constants"
	message "torrent-pi/peerMessage"

	"github.com/jackpal/bencode-go"
)

type ExtensionHandshake struct {
	Map           map[string]int "m"
	Port          int            "p"
	Version       []byte         "v"
	Metadata_size int            "metadata_size" // in bytes
}

func NewExtended(port int) *ExtensionHandshake {
	m := make(map[string]int, 1)
	m["ut_metadata"] = 3
	return &ExtensionHandshake{Map: m, Port: port, Version: []byte(constants.CLIENT_NAME + constants.VERSION)}
}

/* Extension message Headers:
size			description
uint32_t	length prefix. Specifies the number of bytes for the entire message. (big-endian)
uint8_t		bittorrent message ID, = 20
uint8_t		extended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
*/
func (h *ExtensionHandshake) Serialize() *bytes.Reader {
	var b bytes.Buffer
	bencode.Marshal(&b, h)

	buf := make([]byte, 6+b.Len())
	// uint32_t length prefix. Specifies the number of bytes for the entire message. (Big endian)
	binary.BigEndian.PutUint32(buf[0:], uint32(6+b.Len()))
	// uint8_t bittorrent extension message ID = 20
	buf[4] = byte(message.MsgExtended)
	// uint8_t extended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
	buf[5] = byte(0)
	_, err := b.Read(buf[6:])
	if err != nil {
		panic("Bad read of the buffer")
	}

	// Add Payload
	return bytes.NewReader(buf[:])
}

func ReadExtension(r io.Reader) (*ExtensionHandshake, error) {
	fmt.Println("Extension handshake Response!")
	headerBuf := make([]byte, 6)
	_, err := io.ReadFull(r, headerBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := binary.BigEndian.Uint32(headerBuf[:4])
	fmt.Println("Packet length:", pstrlen)
	// Check uint8_t bittorrent message ID == 20
	messageId := uint8(headerBuf[4])
	fmt.Println("messageId:", messageId)

	if messageId != uint8(message.MsgExtended) {
		fmt.Println("bad message Id", messageId)
	}

	handshakeMsgId := uint8(headerBuf[5])
	if handshakeMsgId != 0 {
		fmt.Println("extended message Id does not match 0", handshakeMsgId)
	}
	var handshake ExtensionHandshake
	err = bencode.Unmarshal(r, &handshake)
	if err != nil {
		return nil, err
	}

	return &handshake, err
}
