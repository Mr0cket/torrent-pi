package handshake

import (
	"bytes"
	"fmt"
	"io"
)

/*
The length of the protocol identifier, which is always 19 (0x13 in hex)
The protocol identifier, called the pstr which is always BitTorrent protocol
Eight reserved bytes, all set to 0. We’d flip some of them to 1 to indicate that we support certain extensions. But we don’t, so we’ll keep them at 0.
The infohash that we calculated earlier to identify which file we want
The Peer ID that we made up to identify ourselves
*/

type Handshake struct {
	Pstr       string
	InfoHash   [20]byte
	PeerID     [20]byte
	Extensions [8]byte
}

// New creates a new handshake with the standard pstr
func New(infoHash, peerID [20]byte) *Handshake {
	t := Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
	supportedExtensions := make([]byte, 8)
	// Set support for extension protocol (byte 5 & 0x10)
	supportedExtensions[5] = byte(0x10)
	copy(t.Extensions[:], supportedExtensions) // TODO: indicate support for metadata extension protocol
	return &t
}

func (h *Handshake) Serialize() *bytes.Reader {
	// create the buffer
	buf := make([]byte, len(h.Pstr)+49)
	// length of Protocol string
	buf[0] = byte(len(h.Pstr))

	curr := 1
	// Protocol string
	curr += copy(buf[curr:], h.Pstr)

	// bittorrent protocol extensions string
	curr += copy(buf[curr:], h.Extensions[:])

	curr += copy(buf[curr:], h.InfoHash[:])
	copy(buf[curr:], h.PeerID[:])
	return bytes.NewReader(buf[:])
}

func Read(r io.Reader) (*Handshake, error) {
	var err error
	lengthBuf := make([]byte, 1)

	if _, err := io.ReadFull(r, lengthBuf); err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0]) // length of protocol string
	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, 48+pstrlen)

	if _, err = io.ReadFull(r, handshakeBuf); err != nil {
		return nil, err
	}
	count := pstrlen
	var extensions [8]byte
	count += copy(extensions[:], handshakeBuf[count:count+8])

	var infoHash, peerID [20]byte

	count += copy(infoHash[:], handshakeBuf[count:count+20])
	copy(peerID[:], handshakeBuf[count:count+20])

	h := Handshake{
		Pstr:       string(handshakeBuf[0:pstrlen]),
		InfoHash:   infoHash,
		PeerID:     peerID,
		Extensions: extensions,
	}

	return &h, nil
}
