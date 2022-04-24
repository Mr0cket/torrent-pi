package torrent

import (
	"fmt"
	"os"
	"testing"
)

func TestFromMetadata(t *testing.T) {
	path, err := os.Getwd()
	data, err := os.ReadFile(path + "/../metadata_files/test.torrent")
	if err != nil {
		t.Error(err)
	}

	torrent, err := FromMetadata(data)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(torrent.String())
}
