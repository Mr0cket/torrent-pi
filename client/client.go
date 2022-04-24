package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"torrent-pi/handshake"
	"torrent-pi/peer"
	message "torrent-pi/peerMessage"
)

type ReservedBits [8]byte

type Client struct {
	Conn     net.Conn
	Choked   bool
	peer     peer.Peer
	peerID   [20]byte
	infoHash [20]byte
	Reserved ReservedBits
	handshake.ExtensionHandshake
}

func New(peer peer.Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	// Start bittorrent handshake
	h, err := completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	c := &Client{
		Conn:     conn,
		Choked:   true,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
		Reserved: h.Reserved,
	}

	// Check whether Reserved Bit: 44 (DHT) is set
	if c.Reserved.Has(44) {
		// Client supports extension protocol
		extHandshake, err := completeExtensionHandshake(c.Conn)
		if err == nil {
			c.ExtensionHandshake = *extHandshake
		}
	}
	return c, nil

}

// Read reads and consumes a message from the connection
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*handshake.Handshake, error) {
	fmt.Println("Starting completeHandshake...")
	conn.SetDeadline(time.Now().Add(time.Second * 4))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infohash, peerID)
	if _, err := io.Copy(conn, req.Serialize()); err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("Expected infohash %x but got %x", res.InfoHash, infohash)
	}
	return res, nil
}

func completeExtensionHandshake(conn net.Conn) (h *handshake.ExtensionHandshake, err error) {
	// Check whether there are further messages to be read from the connection
	for msg, err := message.Read(conn); err == nil; msg, err = message.Read(conn) {
		// Handle the message
		if msg.ID == message.MsgExtended {
			// They have initiated extended handshake
			if msg.ExtID != message.ExtHandshake {
				conn.Close()
				return nil, fmt.Errorf("Client doing weird shit")
			}

			// Client sent us extended handshake
			h, err := handshake.ReadExtension(msg.ExtendedMessage.Payload)
			if err != nil {
				conn.Close()
				return nil, err
			}
			// Store the handshake in state
			time.Sleep(time.Second * 5)
			// Send extension handshake to peer
			req := handshake.NewExtended(6881, h.Extensions)
			if _, err := io.Copy(conn, req.Serialize()); err != nil {
				return nil, err
			}
			return h, err
		} else {
			fmt.Printf("Message type %v didn't match extended\n", msg.ID)
		}
	}
	return initateExtensionHandshake(conn)
}

func initateExtensionHandshake(conn net.Conn) (h *handshake.ExtensionHandshake, err error) {
	// Create extension handshake
	req := handshake.NewExtended(6881, message.Map{})

	// Send extension handshake
	if _, err := io.Copy(conn, req.Serialize()); err != nil {
		return nil, err
	}

	// Read any messages from the connection
	for msg, err := message.Read(conn); err != nil; msg, err = message.Read(conn) {
		if err != nil {
			conn.Close()
			return nil, err
		}

		// Check the message type

		if msg.ExtID != message.ExtHandshake {
			conn.Close()
			return nil, fmt.Errorf("Client doing some weird shit")
		}
		// Client sent us extended handshake

		// Read the extended handshake
		h, err = handshake.ReadExtension(msg.Payload)
	}
	return
}

func (r *ReservedBits) Has(bit int) bool {
	byteIndex := bit / 8 // Automatically truncated because bit is an int
	bitIndex := byte(bit % 8)
	return r[byteIndex]&bitIndex != 0
}
