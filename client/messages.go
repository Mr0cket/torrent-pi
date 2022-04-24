package client

import (
	"fmt"
	"math"
	"time"
	message "torrent-pi/peerMessage"
)

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

func (c *ExtensionClient) FetchMetadata() []byte {
	fmt.Println("Fetching metadata...")
	var err error

	// Metadata pieces are in form of 16 KB chunks
	metadataPieces := int(math.Ceil(float64(c.Metadata_size) / float64(message.METADATA_PAYLOAD_SIZE)))
	fmt.Println("total Metadata pieces:", metadataPieces)
	// time.Sleep(time.Second * 2)
	dataBuf := make([]byte, c.Metadata_size)

	// Request the metadata
	// 1. Send a request for the metadata
	for i := 0; i < metadataPieces; i++ {
		var m = &message.Message{}
		c.SendRequestMetadata(i)

		// 2. Wait for the response
		// Throw away all messages until we get a piece
		for m.ID != message.MsgExtended || m.ExtID != message.ExtMsgID(c.Extensions["ut_metadata"]) {
			m, err = message.Read(c.conn)
			if err != nil {
				fmt.Println("Error: ", err)
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
