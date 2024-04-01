package peer

import (
	"fmt"
	"net/url"
)

func (pm PeerManager) Announce(peerChan chan []Peer, port uint16) {
	for _, tracker := range pm.Trackers {
		fmt.Println("Announcing to", tracker.Hostname())
		go func(tracker *url.URL) {
			var peers []Peer
			var err error

			switch tracker.Scheme {
			case "http":
				peers, err = pm.announceHTTP(*tracker, port)
			case "udp":
				peers, err = pm.announceUDP(*tracker, port)
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
}
