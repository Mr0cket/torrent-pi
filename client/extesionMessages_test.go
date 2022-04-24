package client

import (
	"bytes"
	"fmt"
	"testing"
	message "torrent-pi/peerMessage"

	"github.com/jackpal/bencode-go"
)

type MetaDataData struct {
	MsgType int `bencode:"msg_type"`
	Piece   int `bencode:"piece"`
	Size    int `bencode:"total_size"`
}

func TestSendRequestMetadata(t *testing.T) {
	var b bytes.Buffer
	req := MetaDataData{
		MsgType: int(message.MetaRequest),
		Piece:   3,
		Size:    12341234,
	}

	if err := bencode.Marshal(&b, req); err != nil {
		fmt.Println("Error: ", err)
	}
	b.Write([]byte{0x8b, 0xfd, 0x33, 0x87, 0xcf})
	r := bytes.NewReader(b.Bytes())
	data, _ := bencode.Decode(r)
	fmt.Println(data)
}
