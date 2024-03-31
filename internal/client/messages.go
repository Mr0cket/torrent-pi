package client

import (
	"fmt"
	"math"
	"time"
	message "torrent-pi/internal/peerMessage"
)

func (c *Client) SendRequest(pieceIndex, beginByte, length uint) error {
	msg := message.FormatRequest(pieceIndex, beginByte, length)
	_, err := c.Conn.Write(msg.Serialize())
	return err
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

func (c *Client) FetchMetadata() []byte {
	fmt.Println("Fetching metadata...")
	var err error

	// How to find out metadata size?
	// Metadata pieces are in form of 16 KB chunks
	metadataPieces := int(math.Ceil(float64(c.Metadata_size) / float64(message.METADATA_PAYLOAD_SIZE)))
	fmt.Println("total Metadata pieces:", metadataPieces)
	dataBuf := make([]byte, c.Metadata_size)

	// Request the metadata
	// 1. Send a request for the metadata
	for i := 0; i < metadataPieces; i++ {
		var m = &message.Message{}
		msg := message.FormatRequestMetadata(c.Extensions["ut_metadata"], i)
		_, err = c.Conn.Write(msg.Serialize())
		if err != nil {
			fmt.Println("Error: ", err)
		}

		// 2. Wait for the response
		// Throw away all messages until we get a piece
		for m.ID != message.MsgExtended || m.ExtID != message.ExtMsgID(c.Extensions["ut_metadata"]) {
			m, err = message.Read(c.Conn)
			if err != nil {
				fmt.Println("FetchMetadata Error: ", err)
				continue
			}
		}

		// 3. Read the response
		piece, err := message.ParseMetadata(m.ExtendedMessage, c.Extensions)
		time.Sleep(time.Second * 1)
		if err != nil {
			fmt.Println("Error: ", err)
			// Put the piece back to be picked up by another request
			continue
		}

		// 4. Copy the data to the buffer
		pieceOffset := piece.Piece * message.METADATA_PAYLOAD_SIZE
		copy(dataBuf[pieceOffset:], piece.Payload)
	}

	return dataBuf
}
