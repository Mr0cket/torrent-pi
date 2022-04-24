package client

import (
	"bytes"
	"fmt"
	message "torrent-pi/peerMessage"

	"github.com/jackpal/bencode-go"
)

func (c *Client) SendRequestMetadata(startBlock int) {
	fmt.Println("Requesting metadata piece:", startBlock)
	var b bytes.Buffer
	req := message.MetaReq{
		MsgType: int(message.MetaRequest),
		Piece:   startBlock,
	}
	if err := bencode.Marshal(&b, req); err != nil {
		fmt.Println("Error: ", err)
	}

	msg := message.ExtendedMessage{ExtID: message.ExtMsgID(c.Extensions["ut_metadata"]), Payload: b.Bytes()}

	// Send the message
	c.Conn.Write(msg.Serialize())
}
