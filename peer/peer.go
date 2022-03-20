package peer

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

// Unmarshal parses peer IP addresses and ports from a buffer
func Unmarshal(peersBin []byte) ([]Peer, error) {
	const peerSize = 6 // 4 for IP, 2 for port
	numPeers := len(peersBin) / peerSize
	if len(peersBin)%peerSize != 0 {
		err := fmt.Errorf("Received malformed peers")
		return nil, err
	}
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16([]byte(peersBin[offset+4 : offset+6]))
	}
	return peers, nil
}

func (p Peer) String() string {
	return p.IP.String() + ":" + strconv.Itoa(int(p.Port))
}

func RemoveDuplicates(arr []Peer) []Peer {
	keys := make(map[string]bool)
	var uniques []Peer
	for _, peer := range arr {
		if _, value := keys[peer.String()]; !value {
			keys[peer.String()] = true
			uniques = append(uniques, peer)
		}
	}
	return uniques
}
