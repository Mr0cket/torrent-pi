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

type Client struct {
	Conn     net.Conn
	Choked   bool
	peer     peer.Peer
	peerID   [20]byte
	infoHash [20]byte
	ExtensionClient
}

type ExtensionClient struct {
	handshake.ExtensionHandshake
	conn net.Conn
}

func New(peer peer.Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	// Start bittorrent handshake
	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}, nil
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

func NewExtension(peer peer.Peer, peerID, infoHash [20]byte) (*ExtensionClient, error) {
	fmt.Println("NewExtensionClient")
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected with: ", peer.String())
	// Send/recieve bittorrent handshake
	h, err := completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}
	fmt.Println("completed handshake!")

	// TODO check explicitly if "Extension protocol" bit is set using bitwise operations
	// Currently only checking if value of byte which contains the bit is atleast 0x10 (16)
	// This will also pass if any of the first 4 significant bits in the reserved[5] are set
	// Reserved Bit: 44, the fourth most significant bit in the 6th reserved byte i.e. reserved[5] >= 0x10
	if h.Extensions[5] < byte(0x10) {
		conn.Close()
		return nil, fmt.Errorf("Client doesn't support extension handshake")
	}

	// The problem: peer can initialise extended handshake, before we send ours
	// We need to check whether they have already sent us extended handshake before we send ours

	// Check for messages from peer
	// If we receive a message, we know that peer has already sent us extended handshake
	// If we don't receive a message, we know that peer hasn't sent us extended handshake
	// And we can send ours

	// Check whether there are further messages to be read from the connection
	for msg, err := message.Read(conn); err == nil; msg, err = message.Read(conn) {
		// Handle the message
		switch msg.ID {
		case message.MsgExtended:
			// They have already sent us extended handshake
			// We can send ours as a response
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
			return &ExtensionClient{*h, conn}, err
		default:
			fmt.Printf("Message type %v didn't match extended\n", msg.ID)
		}
	}

	// Send/recieve extension handshake
	dict, err := initiateExtensionHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &ExtensionClient{*dict, conn}, err
}

func initiateExtensionHandshake(conn net.Conn) (h *handshake.ExtensionHandshake, err error) {
	fmt.Println("Initiating extension handshake...")
	conn.SetDeadline(time.Now().Add(time.Second * 4))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	// Create extension handshake
	req := handshake.NewExtended(6881, message.Map{})

	// Send extension handshake
	if _, err := io.Copy(conn, req.Serialize()); err != nil {
		return nil, err
	}

	// Read any messages from the connection
	for msg, err := message.Read(conn); msg.ID != message.MsgExtended; msg, err = message.Read(conn) {
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
