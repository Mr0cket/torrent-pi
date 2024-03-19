package handshake

import (
	"bytes"
	"fmt"
	"torrent-pi/internal/constants"
	message "torrent-pi/internal/peerMessage"

	"github.com/jackpal/bencode-go"
)

// Set our supported extensions and default identifiers
var supportedExtensions = message.Map{
	"ut_metadata": 2,
}

type ExtensionHandshake struct {
	Extensions    message.Map `bencode:"m"`
	Port          int         `bencode:"p"`             // Port to connect to
	Version       string      `bencode:"v"`             // Verson of the peer's client
	Metadata_size int         `bencode:"metadata_size"` // in bytes
	MyIP          []byte      `bencode:"yourip"`        // My IP (as seen by the other peer)
	// reqLimit      []int  "reqq"          // Request queue size limit
}

func NewExtended(port int, extensions message.Map) *ExtensionHandshake {
	m := supportedExtensions
	for extension, extensionID := range extensions {
		if supportedExtensions[extension] > 0 && extensionID > 0 {
			m[extension] = extensionID
		}
	}
	return &ExtensionHandshake{Extensions: m, Port: port, Version: string(constants.CLIENT_NAME + constants.VERSION)}
}

/*
	Extension message Headers:

size			description
uint32_t	length prefix. Specifies the number of bytes for the entire message. (big-endian)
uint8_t		bittorrent message ID, = 20
uint8_t		extended message ID. 0 = handshake, >0 = extended message as specified by the handshake.
*/
func (h *ExtensionHandshake) Serialize() *bytes.Reader {
	var b bytes.Buffer
	if err := bencode.Marshal(&b, *h); err != nil {
		fmt.Println("Error bencoding Extended handshake:", err)
	}

	req := message.ExtendedMessage{ExtID: 0, Payload: b.Bytes()}
	return bytes.NewReader(req.Serialize())
}

func ReadExtension(buf []byte) (*ExtensionHandshake, error) {
	r := bytes.NewReader(buf)
	var handshake ExtensionHandshake
	err := bencode.Unmarshal(r, &handshake)
	if err != nil {
		return nil, err
	}
	fmt.Println("Extension handshake recieved: ")
	fmt.Println("Extensions:", handshake.Extensions)
	fmt.Println("Metadata_size:", handshake.Metadata_size)
	fmt.Println("Version:", handshake.Version)
	fmt.Println("Port:", handshake.Port)
	fmt.Println("MyIP:", handshake.MyIP)

	return &handshake, err
}
