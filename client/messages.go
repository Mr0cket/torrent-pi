package client

import (
	"encoding/binary"
	"fmt"
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

func (c *MetadataClient) FetchMetadata() []byte {
	fmt.Println("Fetching metadata...")
	// Metadata pieces are in form of 16 byte chunks
	metadataPieces := c.Metadata_size / 16
	fmt.Println("Metadata pieces: ", metadataPieces)
	buf := make([]byte, c.Metadata_size)

	// size of metadata piece as a fraction (256th) of the total size
	// e.g size 0 = 1/256 * metdata_size
	piece_size := 0

	// Request the metadata
	// 1. Send a request for the metadata
	c.SendRequestMetadata(1, piece_size)

	// 2. Read the response
	m, err := message.ReadExtension(c.conn)
	if err != nil {
		fmt.Println("Error reading metadata: ", err)
	}
	fmt.Println("Payload: ", m.Payload)

	// 3. Parse the response
	if m.ID != c.Map["ut_metadata"] {
		fmt.Printf("Error: Bad extended message ID expected %v to match: %v", m.ID, c.Map["ut_metadata"])
	}
	if m.Payload[0] != 1 {
		fmt.Println("Error: Expected metadata message ID 1")
	}
	totalSize := binary.BigEndian.Uint32(m.Payload[1:5])

	fmt.Println("Total metadata size from req:", totalSize)

	offset := binary.BigEndian.Uint32(m.Payload[5:9])
	fmt.Println("Offset: ", offset)

	// 4. Append the response to the metadata
	copy(buf[offset:], m.Payload[9:])
	// 5. Repeat until all pieces are received
	// }
	fmt.Println("Got metadata: ", buf)
	return buf
}

func (c *MetadataClient) SendRequestMetadata(startBlock, size int) {
	// append payload headers
	buf := make([]byte, 3)
	buf[0] = byte(0)          // metadata message type (0 = request), 1 = response  (not implemented), 2 = reject  (not implemented)
	buf[1] = byte(startBlock) // the piece index to request
	buf[2] = byte(size)       // the size of the block to request
	msg := message.ExtendedMessage{ID: c.Map["ut_metadata"], Payload: buf[:]}
	// Send the message
	c.conn.Write(msg.Serialize())
}
