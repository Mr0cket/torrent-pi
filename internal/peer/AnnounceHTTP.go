package peer

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
)

const protocol_id uint64 = 0x41727101980

type HTTPTrackerResponse struct {
	interval   int
	incomplete int
	complete   int
	downloaded int
	peers      []byte
}

func (pm *PeerManager) announceHTTP(tracker url.URL, port uint16) (peers []Peer, err error) {
	// Build the tracker URL
	trackerURL, err := pm.buildTrackerURL(tracker, port)
	if err != nil {
		fmt.Println("Error building tracker URL:", err)
		return
	}
	// Build a client to talk to the tracker
	trackerClient := http.Client{Timeout: time.Second * 15}
	res, err := trackerClient.Get(trackerURL)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	// decode the bEncode response
	data := HTTPTrackerResponse{}
	err = bencode.Unmarshal(res.Body, &data)
	// Unmarshall the peers
	peers, err = Unmarshal(data.peers)
	return
}

func (pm *PeerManager) buildTrackerURL(trackerURL url.URL, port uint16) (string, error) {
	params := url.Values{
		"info_hash":  []string{string(pm.InfoHash[:])},
		"peer_id":    []string{string(pm.PeerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
	}
	trackerURL.RawQuery = params.Encode()
	return trackerURL.String(), nil
}
