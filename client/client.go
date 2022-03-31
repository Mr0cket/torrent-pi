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
}

type MetadataClient struct {
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
	_, err := io.Copy(conn, req.Serialize())
	if err != nil {
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

func (c *Client) SendRequest(pieceIndex, beginByte, length int) error {
	req := message.FormatRequest(pieceIndex, beginByte, length)
	c.Conn.Write(req.Serialize())
	return nil
}

// Send unchoke message
func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func NewExtensionClient(peer peer.Peer, peerID, infoHash [20]byte) (*MetadataClient, error) {
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
	fmt.Println("Supported extensions:", h.Extensions)
	// Send/recieve extension handshake
	_, err = completeExtensionHandshake(conn)

	return &MetadataClient{conn}, err
}

func completeExtensionHandshake(conn net.Conn) ([]byte, error) {
	fmt.Println("Starting extension handshake...")
	conn.SetDeadline(time.Now().Add(time.Second * 4))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	// Create extension handshake
	req := handshake.NewExtended()
	_, err := io.Copy(conn, req.Serialize())
	if err != nil {
		return nil, err
	}
	h, err := handshake.ReadExtension(conn)
	if err != nil {
		fmt.Println("Read Extension error:", err.Error())
	}
	fmt.Println(h)

	buf := make([]byte, 0)
	return buf, err
}

func (c *MetadataClient) GetMetadata() {
	fmt.Println("TODO: Get metadata")
}
