package torrent

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"torrentBox/lib"
	"torrentBox/peer"
)

// Implement announce UDP request
func (t *Torrent) announceUDP(tracker url.URL, port uint16) (peers []peer.Peer, err error) {

	// Implement UDP client that sends a UDP packet to the tracker
	// and waits for a response.

	/* Step 1: Connect */

	// Generate transaction ID
	transactionID := rand.Uint32()

	// Create the UDP packet
	packet := make([]byte, 16)

	// Write the protocol ID
	binary.BigEndian.PutUint64(packet[0:8], protocol_id)
	// Write the action
	binary.BigEndian.PutUint32(packet[8:12], uint32(0))
	// Write the transaction ID
	binary.BigEndian.PutUint32(packet[12:16], transactionID)

	reader := bytes.NewReader(packet)

	// Send the UDP packet
	res, err := lib.UDPRequest(tracker.Host, reader)
	if err != nil {
		fmt.Println("Error reading UDP tracker:", err)
		return
	}

	// Parse the response
	res_action := binary.BigEndian.Uint32(res[0:4])
	res_transactionID := binary.BigEndian.Uint32(res[4:8])
	connectionID := binary.BigEndian.Uint64(res[8:16])

	if res_action != 0 {
		return peers, fmt.Errorf("Error: action not 0")
	}

	if res_transactionID != transactionID {
		return peers, fmt.Errorf("Error: transaction ID not equal")
	}

	fmt.Println("Successfully connected to UDP tracker")

	/* Step 2: Announce */

	// Create Announce UDP packet
	packet = make([]byte, 98)
	transactionID = rand.Uint32()

	binary.BigEndian.PutUint64(packet[0:8], connectionID)
	// action
	binary.BigEndian.PutUint32(packet[8:12], 0x1)
	// transactionID
	binary.BigEndian.PutUint32(packet[12:16], transactionID)
	// info hash
	copy(packet[16:36], t.InfoHash)
	// peer id
	copy(packet[36:56], t.PeerID)
	// downloaded
	binary.BigEndian.PutUint64(packet[56:64], t.Downloaded)
	// left
	binary.BigEndian.PutUint64(packet[64:72], ^uint64(0))
	// uploaded
	binary.BigEndian.PutUint64(packet[72:80], 0)
	// event
	binary.BigEndian.PutUint32(packet[80:84], 2)
	// IP address
	binary.BigEndian.PutUint32(packet[84:88], 0)
	// key
	binary.BigEndian.PutUint32(packet[88:92], 0)
	// num want
	binary.BigEndian.PutUint32(packet[92:96], rand.Uint32())
	// port
	binary.BigEndian.PutUint16(packet[96:98], uint16(port))

	if tracker.Path != "" || tracker.RawQuery != "" {
		// Add URLData extension to packet
		URLData := make([]byte, 2+len(tracker.Path)+len(tracker.RawQuery))
		// Add Option-type
		URLData[0] = byte(0x2)
		// Add URLData length
		URLData[1] = byte(len(tracker.Path) + len(tracker.RawQuery))
		// Add URLData
		var data []byte
		if tracker.Path != "" {
			data = []byte(tracker.Path)
		}
		if tracker.RawQuery != "" {
			data = append(data, []byte("?"+tracker.RawQuery)...)
		}

		copy(URLData[2:], data)
		// Reprovision packet to include URLData
		packet = append(packet, URLData...)
	}

	fmt.Print("Announce packet length:", len(packet))
	// Send the UDP packet
	res, err = lib.UDPRequest(tracker.Host, bytes.NewReader(packet))
	if err != nil {
		fmt.Println("Error reading UDP tracker:", err)
		return
	}

	// Verify & parse the response
	res_action = binary.BigEndian.Uint32(res[0:4])
	res_transactionID = binary.BigEndian.Uint32(res[4:8])
	interval := binary.BigEndian.Uint32(res[8:12])
	leechers := binary.BigEndian.Uint32(res[12:16])
	seeders := binary.BigEndian.Uint32(res[16:20])

	if res_action != 1 {
		return peers, fmt.Errorf("Error: action not Announce")
	}
	if res_transactionID != transactionID {
		return peers, fmt.Errorf("Error: transaction ID not equal")
	}

	peers = make([]peer.Peer, leechers+seeders)
	fmt.Println("Interval", interval)
	fmt.Println("Leechers:", leechers)
	fmt.Println("Seeders:", seeders)

	// Get the peer ip addresses etc.
	for i := 0; i < len(peers); i++ {
		startIP := 20 + i*6
		peers[i].IP = net.IPv4(res[startIP], res[startIP+1], res[startIP+2], res[startIP+3])
		peers[i].Port = binary.BigEndian.Uint16(res[startIP+4 : startIP+6])
	}
	fmt.Println("Peers:", peers)
	return
}
