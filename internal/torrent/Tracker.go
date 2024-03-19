package torrent

import (
	"fmt"
	"net/url"
	"sync"
	"torrent-pi/internal/peer"
)

// Announce to all trackers
func (t Torrent) AnnounceAll(port uint16) []peer.Peer {
	fmt.Println("Announcing to all trackers")
	var peerChan = make(chan []peer.Peer, len(t.Trackers))
	var peers []peer.Peer
	defer close(peerChan)
	wg := sync.WaitGroup{}

	for i, tracker := range t.Trackers {
		wg.Add(i)
		fmt.Println("Announcing to", tracker.Hostname())
		go func(tracker *url.URL) {
			defer wg.Done()
			var peers []peer.Peer
			var err error

			switch tracker.Scheme {
			case "http":
				peers, err = t.announceHTTP(tracker, port)
			case "udp":
				peers, err = t.announceUDP(*tracker, port)
			default:
				fmt.Println("unsupporter tracker scheme:", tracker.Scheme)
				return
			}
			if err != nil {
				fmt.Println("Error announcing", tracker.Hostname(), err)
				return
			}
			peerChan <- peers
		}(tracker)
	}

	// Wait for trackers to finish
	wg.Wait()

	for newPeers := range peerChan {
		peers = append(peers, newPeers...)
	}

	fmt.Println("Amount of peers", len(peers))
	return peer.RemoveDuplicates(peers)
}

func (t Torrent) AnnounceRace(port uint16) []peer.Peer {
	var peerChan = make(chan []peer.Peer, 1)
	defer close(peerChan)
	once := sync.Once{}

	for _, tracker := range t.Trackers {
		go func(tracker *url.URL) {
			var peers []peer.Peer
			var err error

			switch tracker.Scheme {
			case "http":
				peers, err = t.announceHTTP(tracker, port)
			case "udp":
				peers, err = t.announceUDP(*tracker, port)
			default:
				fmt.Println("unsupported tracker scheme:", tracker.Scheme)
				return
			}
			if err != nil {
				fmt.Println("Error announcing", tracker.Hostname(), err)
				return
			}
			once.Do(func() { peerChan <- peers })
		}(tracker)
	}

	return <-peerChan
}
