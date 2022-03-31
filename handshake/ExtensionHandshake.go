package handshake

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	message "torrent-pi/peerMessage"

	"github.com/jackpal/bencode-go"
)

type dict struct {
	m []interface{}
	// metadata_size int
}
type MetadataExtension struct {
	dict
	// p int // Local TCP listening port
}

type AnotherExtension struct {
	dict struct {
		a string
		b int
	}
}

type ExtensionHandshake MetadataExtension

func NewExtended() *ExtensionHandshake {
	// m := make(map[string]int, 1)
	// m["ut_metadata"] = 3
	arr := []interface{}{"ut_metadata", 3}
	return &ExtensionHandshake{
		dict: dict{m: arr},
	}
}

// TODO: Create the extension message & handshake http://www.bittorrent.org/beps/bep_0010.html
func (h *ExtensionHandshake) Serialize() *bytes.Reader {
	var b bytes.Buffer
	bencode.Marshal(&b, h.dict)

	buf := make([]byte, 6+b.Len())
	// Headers - 6 bytes
	// uint32_t length prefix. Specifies the number of bytes for the entire message. (Big endian)
	binary.BigEndian.PutUint32(buf[0:], uint32(6+b.Len()))
	// uint8_t bittorrent message ID, = 20
	buf[4] = byte(message.MsgExtended)
	// uint8_t xtended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
	buf[5] = byte(0)
	_, err := b.Read(buf[6:])
	if err != nil {
		panic("Bad read of the buffer")
	}

	// Payload
	fmt.Println("Message length: ", len(buf))
	fmt.Println("byte:", string(buf[6:]))
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

	data, err := bencode.Decode(r)
	fmt.Println("bencode Decoded message:")
	fmt.Println(data)
	return &ExtensionHandshake{}, err
}
