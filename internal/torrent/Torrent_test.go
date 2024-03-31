package torrent

import (
	"fmt"
	"os"
	"testing"
)

func TestFromMetadata(t *testing.T) {
	path, _ := os.Getwd()
	data, err := os.ReadFile(path + "/../../samples/lord_of_the_rings.torrent")
	if err != nil {
		t.Error(err)
	}

	torrent, err := FromMetadata(data)
	if err != nil {
		t.Error(err)
	}
	if torrent.Name == "" {
		t.Error("No Torrent name")
	}

	fmt.Println("Torrent name:", torrent.Name)
	fmt.Println(torrent.String())
}
