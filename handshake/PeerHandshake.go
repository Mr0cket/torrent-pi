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
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

// New creates a new handshake with the standard pstr
func New(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() *bytes.Reader {
	// create the buffer
	buf := make([]byte, len(h.Pstr)+49)
	// length of Protocol string
	buf[0] = byte(len(h.Pstr))

	curr := 1
	// Protocol string
	curr += copy(buf[curr:], h.Pstr)
	// Empty string for bittorrent protocol extensions
	curr += copy(buf[curr:], make([]byte, 8))
	curr += copy(buf[curr:], h.InfoHash[:])
	copy(buf[curr:], h.PeerID[:])
	return bytes.NewReader(buf[:])
}

func Read(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	pstrlen := int(lengthBuf[0])
	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}

	handshakeBuf := make([]byte, 48+pstrlen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var infoHash, peerID [20]byte

	copy(infoHash[:], handshakeBuf[pstrlen+8:pstrlen+8+20])
	copy(peerID[:], handshakeBuf[pstrlen+8+20:])

	h := Handshake{
		Pstr:     string(handshakeBuf[0:pstrlen]),
		InfoHash: infoHash,
		PeerID:   peerID,
	}

	return &h, nil
}
