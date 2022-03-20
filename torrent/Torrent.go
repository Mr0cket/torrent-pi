package torrent

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"torrentBox/peer"
)

type Torrent struct {
	PeerID     []byte
	InfoHash   []byte
	Name       string
	Trackers   []*url.URL
	Downloaded uint64
	Peers      []peer.Peer
	length     int
}

const MAX_PORT = 65535

// Construct Torrent from magnet URL
func NewTorrent(magnetURL *url.URL) (Torrent, error) {
	var err error

	// Parse the magnet link
	var trackers = magnetURL.Query()["tr"]
	var name = magnetURL.Query().Get("dn")
	var infoHash_hex = strings.TrimPrefix(magnetURL.Query().Get("xt"), "urn:btih:")
	infoHash, err := hex.DecodeString(infoHash_hex)
	// TODO: Validate magnet link & info hash

	// parse trackers
	Trackers := make([]*url.URL, len(trackers))
	for i, t := range trackers {
		tracker, err := url.Parse(t)
		if err != nil {
			continue
		}
		Trackers[i] = tracker
	}

	// Generate peerID
	baseID := "MwowClient1.0_"
	peerID := baseID + strconv.Itoa(int(time.Now().Unix()))[:20-len(baseID)]

	t := Torrent{
		PeerID:   []byte(peerID),
		InfoHash: infoHash,
		Name:     name,
		Trackers: Trackers,
		Peers:    make([]peer.Peer, 0),
	}

	return t, err
}

/* Torrent Methods */
func (t Torrent) Download() {
	fmt.Println("Downloading", t.Name)

	// Announce to all trackers
	var port uint16 = 6881
	peers := t.AnnounceAll(port)

	fmt.Printf("Announce successful!!\nFound %d Peers: %v", len(peers), peers)

	// TODO: Download from peers

}
