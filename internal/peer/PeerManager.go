package peer

import (
	"fmt"
	"net/url"
	"sync"
)

// Manages peers
// What needs to happen?
// there needs to be a process to manage the pool of peers.
// this manager runs in a separate thread and stores a pool of peers.
// peers can have status 'good' 'bad' indicating whether they work or not.

// The peer manager polls (working) trackers at regular intervals to retrieve new potential peers.

// A Peer is tested by sending a packet.
// request timeout: peer status -> bad
// request success: peer status -> good

// When downloading torrent, call peer manager for a peer (or list of peers).
// packet req timeout: -> peer status -> bad
// packet req success: -> peer status -> good

type PeerStatus int

const UNKNOWN PeerStatus = 0
const GOOD PeerStatus = 1
const BAD PeerStatus = 2

var lock sync.Mutex

type PeerState struct {
	peer   Peer
	status PeerStatus
	conns  int // connections (perhaps not needed)
}
type TorrentStats struct {
	Downloaded uint64
	Uploaded   uint64
}

type PeerManager struct {
	peers    map[string]PeerState
	InfoHash []byte
	PeerID   []byte
	Trackers []*url.URL
}

func NewPeerManager(infoHash, peerId []byte, trackers []*url.URL) PeerManager {
	pm := PeerManager{InfoHash: infoHash, PeerID: peerId, Trackers: trackers, peers: make(map[string]PeerState, 0)}

	lock.Lock()
	return pm
}

func (pm PeerManager) GetPeers() []Peer {
	peers := make([]Peer, len(pm.peers))
	for _, peer := range pm.peers {
		if len(peer.peer.IP) > 0 {
			peers = append(peers, peer.peer)
		}
	}
	return peers
}

func (pm PeerManager) GetPeer() Peer {
	var p Peer
	for _, peer := range pm.peers {
		if peer.status == BAD {
			continue
		}
		p = peer.peer
	}
	return p
}

func (pm PeerManager) AddPeers(peers []Peer) {
	for _, peer := range peers {
		// Skip if peer exists
		if _, ok := pm.peers[peer.IP.String()]; ok || peer.IP == nil {
			continue
		}
		pm.peers[peer.IP.String()] = PeerState{peer: peer, conns: 0}
	}
	// Remove any nil keys
	for key := range pm.peers {
		fmt.Println("Peer key", key)
	}
}

func (pm PeerManager) SetPeerStatus(peerIp string, status PeerStatus) {
	if _, ok := pm.peers[peerIp]; ok {
		temp := pm.peers[peerIp]
		temp.status = status
		pm.peers[peerIp] = temp
		fmt.Printf("Peer %v status changed to %v\n", peerIp, pm.peers[peerIp].status)
	} else {
		fmt.Println("Err: Unable to find peer", peerIp)
	}
}

func (pm PeerManager) DropPeer(peerIp string) {
	if _, ok := pm.peers[peerIp]; ok {
		temp := pm.peers[peerIp]
		temp.conns = 0
		pm.peers[peerIp] = temp
	}
}

func (pm PeerManager) Start(port uint16) {
	fmt.Println("Announcing to all trackers")

	var peerChan = make(chan []Peer, len(pm.Trackers))
	defer close(peerChan)

	pm.Announce(peerChan, port)

	for newPeers := range peerChan {

		if len(pm.peers) < 1 && len(newPeers) > 0 {
			pm.AddPeers(newPeers)
			lock.Unlock()
		} else {
			pm.AddPeers(newPeers)
		}
	}
	// TODO announce to tracker at regular intervals, following the value of 'interval' in response
}

// Wait until the peermanager has found at least 1 peer before proceeding with execution
func (pm PeerManager) WaitReady() {

	// Wait for the mutex to become unlocked
	fmt.Println("Waiting for peer Lock...")
	lock.Lock()
	fmt.Println("Peer Lock unlocked")
}
