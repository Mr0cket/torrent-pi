package handshake

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jackpal/bencode-go"
)

type MetaDataData struct {
	Msg        int `bencode:"msg_type"`
	Piece      int `bencode:"piece"`
	Total_size int `bencode:"total_size"`
}

func TestSerialize(t *testing.T) {
	meta := MetaDataData{
		Msg:        1,
		Piece:      0,
		Total_size: 16350,
	}
	var b bytes.Buffer
	bencode.Marshal(&b, meta)
	b.Write([]byte("abasdfasdf qwerqwerasdf qwefasdfafbl;qeg vaspvg[]p="))
	fmt.Println(b.Bytes())
	r := bytes.NewReader(b.Bytes())
	var m MetaDataData
	bencode.Unmarshal(r, &m)
	fmt.Println("type:", m.Msg, "piece:", m.Piece, "size:", m.Total_size)
}
