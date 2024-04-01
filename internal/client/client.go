package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"torrent-pi/internal/constants"
	"torrent-pi/internal/handshake"
	"torrent-pi/internal/peer"
	message "torrent-pi/internal/peerMessage"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	peer     peer.Peer
	peerID   [20]byte
	infoHash [20]byte
	Reserved ReservedBits
	Bitfield message.Bitfield
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
	fmt.Println("client reservedBits", c.Reserved.String())
	if !c.Reserved.Has(44) {
		return nil, errors.New("doesn't support extension bit")
	} else {
		fmt.Println("Client supports extension bit")
	}

	// Client supports extension protocol
	fmt.Println("Starting completeExtensionHandshake")
	extHandshake, err := completeExtensionHandshake(c.Conn)

	if err != nil {
		return nil, err
	}
	fmt.Println("Finished completeExtensionHandshake")
	c.ExtensionHandshake = *extHandshake

	return c, nil

}

// Read reads and consumes a message from the connection
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	if err == nil && msg != nil {
		fmt.Println(c.peer.IP, "->", msg.TypeString())
	}
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
		fmt.Println("Error reading handshake to conn", err)
		return nil, err
	}
	fmt.Println("Read handshake to conn", conn.RemoteAddr())

	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infohash)
	}
	return res, nil
}

func completeExtensionHandshake(conn net.Conn) (h *handshake.ExtensionHandshake, err error) {
	// Check whether there are further messages to be read from the connection
	for msg, err := message.Read(conn); err == nil; msg, err = message.Read(conn) {
		fmt.Println("Recieved message:", msg.TypeString())
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
		fmt.Println("Recieved message:", msg.TypeString())

		// Check the message type

		if msg.ExtID != message.ExtHandshake {
			conn.Close()
			return nil, fmt.Errorf("Client doing some weird shit")
		}
		// Client sent us extended handshake

		// Read the extended handshake
		h, err = handshake.ReadExtension(msg.Payload)
		if err != nil {
			fmt.Println("Extension handshake failed to read extension msg")
		}
	}
	return
}

// Download a full piece by sending consecutive requests for blocks which make up that piece
func (c Client) DownloadPiece(pieceBuffer []byte, pieceIndex, blockCount uint) error {
	errorCount := 0
	var err error
	for blockIndex := range make([]int, blockCount) {
		var msg = &message.Message{}
		startByte := uint(blockIndex) * constants.BLOCK_SIZE
		for {
			// fmt.Printf("Piece #%d block #%d -> %v\n", pieceIndex, blockIndex, c.peer.IP)

			// Throw away all messages until we get a piece
			c.SendRequest(pieceIndex, startByte, constants.BLOCK_SIZE)
			if msg == nil {
				msg = &message.Message{}
			}
			for msg.ID != message.MsgPiece {
				msg, err = c.Read()
				if err != nil {
					errorCount++
					if errorCount > 3 {
						return err
					}
					fmt.Println("Error: ", err)
					break
				}
			}
			if msg.ID == message.MsgPiece {
				_, err := message.ParsePiece(int(pieceIndex), pieceBuffer, msg)
				if err != nil {
					fmt.Println("Buffer error", err)
					return err
				}
				break
			}
		}
	}
	return nil
}
