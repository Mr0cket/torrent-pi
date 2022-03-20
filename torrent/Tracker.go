package torrent

import (
	"fmt"
	"net/url"
	"torrentBox/peer"
)

// Announce to all trackers
func (t Torrent) AnnounceAll(port uint16) []peer.Peer {
	fmt.Println("Announcing to all trackers")
	var trackerCounter = 0
	var p = make(chan []peer.Peer, len(t.Trackers))
	var done = make(chan bool)
	var newPeers []peer.Peer
	var allPeers []peer.Peer

	defer close(p)
	for _, tracker := range t.Trackers {
		go t.Announce(p, done, tracker, port)
	}
	for trackerCounter < len(t.Trackers) {
		select {
		case newPeers = <-p:
		case <-done:
			allPeers = append(allPeers, newPeers...)
			trackerCounter++
		}
	}

	return peer.RemoveDuplicates(allPeers)
}

func (t *Torrent) Announce(newPeers chan []peer.Peer, done chan bool, tracker *url.URL, port uint16) (peers []peer.Peer, err error) {
	defer func() { done <- true }()
	fmt.Println("Announcing to", tracker.Hostname())
	// println()
	if tracker.Scheme == "http" {
		peers, err = t.announceHTTP(tracker, port)
	} else if tracker.Scheme == "udp" {
		peers, err = t.announceUDP(*tracker, port)
	}
	if err != nil {
		fmt.Println("Error announcing", tracker.Hostname(), err)
		return
	}
	newPeers <- peers
	return
}
