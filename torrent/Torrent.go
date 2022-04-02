package torrent

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"torrent-pi/client"
	"torrent-pi/constants"
	"torrent-pi/peer"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

type PeerID [20]byte

type Torrent struct {
	PeerID      [20]byte
	InfoHash    [20]byte
	Name        string
	Trackers    []*url.URL
	Downloaded  uint64
	Peers       []peer.Peer
	PieceHashes [][20]byte
	PieceLength int
	length      int
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

	t := Torrent{
		Name:     name,
		Trackers: Trackers,
		Peers:    make([]peer.Peer, 0),
	}
	copy(t.PeerID[:], []byte(constants.PEER_ID))
	copy(t.InfoHash[:], infoHash)

	// TODO Retrieve torrent metadata from the "swarm"... http://www.bittorrent.org/beps/bep_0009.html
	peers := t.AnnounceRace(8081)
	fmt.Println("Got peers")
	fmt.Println(peers)

	// Retrieve file metadata with metadata extension protocol
	for _, peer := range peers {
		client, err := client.NewExtension(peer, t.PeerID, t.InfoHash)
		if err != nil {
			fmt.Println("Error connecting to peer", err.Error())
			continue
		}
		client.FetchMetadata()
		break
	}

	return t, err
}

/* Torrent Methods */
func (t Torrent) Download() {
	fmt.Println("Downloading", t.Name)
	// // Announce to all trackers
	var port uint16 = 6881
	peers := t.AnnounceAll(port)

	fmt.Printf("Announce successful!!\nFound %d Peers: %v\n", len(peers), peers)

	// 1. connect to Peer
	// 2. send handshake
	// 3. send bitfield
	// 4. send request
	// 5. send piece
	// 6. send cancel
	// 7. send port
	// 8. send keep-alive
	// 9. send choke
	// 10. send unchoke
	// 11. send interested
	// 12. send not interested
	// 13. send have

	// TODO: Download File

	// Init queues for workers to retrieve work and send results

	// Write buffer to file

}
