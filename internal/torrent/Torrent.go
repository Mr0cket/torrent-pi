package torrent

import (
	"bytes"
	"container/heap"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"torrent-pi/internal/client"
	"torrent-pi/internal/constants"
	"torrent-pi/internal/lib"
	"torrent-pi/internal/peer"
	message "torrent-pi/internal/peerMessage"
	"torrent-pi/internal/utils"

	"github.com/jackpal/bencode-go"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384

// MaxBacklog is the number of unfulfilled requests a client can have in its pipeline
const MaxBacklog = 5

type PeerID [20]byte

// There is a key 'length' or a key 'files', but not both or neither.
// If length is present then the download represents a single file
// otherwise it represents a set of files which go in a directory structure.

type Trackers []*url.URL

func (t Trackers) String() []string {
	strs := make([]string, len(t))

	for _, tracker := range t {
		strs = append(strs, tracker.String())
	}
	return strs
}

type Torrent struct {
	PeerID   [20]byte `bencode:"-"`
	InfoHash [20]byte `bencode:"info_hash"`
	Name     string   `bencode:"name"`
	Trackers Trackers `bencode:"announce_list"`

	// pieces maps to a string whose length is a multiple of 20.
	// It is to be subdivided into strings of length 20, each of which is the SHA1 hash of the piece at the corresponding index.
	PieceHashesString string   `bencode:"pieces"`
	PieceHashes       [][]byte `bencode:"-"`
	PieceLength       uint     `bencode:"piece length"`
	Length            uint     `bencode:"length"`

	// For the purposes of the other keys, the multi-file case is treated as only having a single file
	// by concatenating the files in the order they appear in the files list.
	Files       Files            `bencode:"files"`
	Downloaded  uint64           `bencode:"-"`
	PeerManager peer.PeerManager `bencode:"-"`
}

const MAX_PORT = 65535

// Construct a Torrent from magnet URL
func NewTorrentFromMagnet(magnetURL *url.URL) (Torrent, error) {
	var err error

	var trackerUrls = magnetURL.Query()["tr"]
	var name = magnetURL.Query().Get("dn")
	var infoHash_hex = strings.TrimPrefix(magnetURL.Query().Get("xt"), "urn:btih:")
	infoHash, err := hex.DecodeString(infoHash_hex)
	// TODO: Validate magnet link & info hash

	// parse trackers
	trackers := make([]*url.URL, len(trackerUrls))
	for i, t := range trackerUrls {
		tracker, err := url.Parse(t)
		if err != nil {
			continue
		}
		trackers[i] = tracker
	}

	t := Torrent{
		Name:        name,
		Trackers:    trackers,
		PeerManager: peer.NewPeerManager(infoHash, []byte(constants.PEER_ID), trackers),
	}
	copy(t.PeerID[:], []byte(constants.PEER_ID))
	copy(t.InfoHash[:], infoHash)

	// TODO Retrieve torrent metadata from the "swarm"... http://www.bittorrent.org/beps/bep_0009.html

	// Start PeerManager which polls/updates trackers at intervals
	go t.PeerManager.Start(6881)

	t.PeerManager.WaitReady()
	fmt.Println("Peers in Torrent module", t.PeerManager.GetPeers())

	// Retrieve file metadata with metadata extension protocol
	for _, peer := range t.PeerManager.GetPeers() {
		c, err := client.New(peer, t.PeerID, t.InfoHash)
		if err != nil {
			// fmt.Println("Error connecting to peer:", err)
			continue
		}
		defer c.Conn.Close()

		metadata := c.FetchMetadata()

		r := bytes.NewReader(metadata)
		bencode.Unmarshal(r, &t)
		break
	}
	t.PieceHashes = utils.SplitStringToBytes(t.PieceHashesString, 20)

	return t, err
}

/* Torrent Methods */
func (t Torrent) Download() {
	fmt.Println("Downloading", t.Name)
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
	// Write the buffer to file at regular intervals

	// Okay, we have the piece hashes, now lets start some workers to fetch pieces incrementally
	// For now, find the .mp4 file

	var fileToDownload File

	fmt.Println("##### Files #####")
	if t.Length > 0 {
		fileToDownload = File{
			Path:   []string{t.Name},
			Length: int(t.Length),
		}
		fmt.Println(fileToDownload.String())
	}
	for _, file := range t.Files {
		fmt.Println(file.String())
		if strings.HasSuffix(file.Path[0], "mp4") || strings.HasSuffix(file.Path[0], ".mkv") {
			fileToDownload = file
			fmt.Println("Found media to download ", fileToDownload.String())
			break
		}
	}

	// Calculate the piece range for a file in a .torrent distribution

	var startByte uint

	// Find file startPiece by summing length of all files with file index smaller than target file
	for _, file := range t.Files {
		if file.String() == fileToDownload.String() {
			break
		}
		startByte += uint(file.Length)
	}
	startPiece := startByte / t.PieceLength
	endPiece := (startByte+uint(fileToDownload.Length))/t.PieceLength - 1
	blockCount := t.PieceLength / constants.BLOCK_SIZE

	fmt.Printf("startPiece: %d, endPiece: %d\n", startPiece, endPiece)
	connections := make([]client.Client, 0)
	piecesQueue := make(lib.PriorityQueue, (endPiece - startPiece))
	i := 0
	for pieceIndex := startPiece; pieceIndex < endPiece; pieceIndex++ {
		piecesQueue[i] = &lib.Item{
			Value:    pieceIndex,
			Priority: max(int(endPiece-pieceIndex), 1),
			Index:    i,
		}
		i++
	}
	heap.Init(&piecesQueue)

	max_connections := 10
	wg := sync.WaitGroup{}
	wg.Add(1)
	fileLock := sync.Mutex{}
	f, err := os.OpenFile("downloads/"+string(fileToDownload.Path[0]), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file", fileToDownload.Path[0], err)
	}

	go func() {
		for len(connections) < max_connections {
			p := t.PeerManager.GetPeer()
			if len(p.IP) == 0 {
				fmt.Println("No new peers, breaking loop")
				break
			}
			fmt.Printf("Peer Connection %s -> starting \n", p.String())
			defer t.PeerManager.DropPeer(p.IP.String())

			c, err := client.New(p, t.PeerID, t.InfoHash)
			if err != nil {
				fmt.Println(err)
				// t.PeerManager.SetPeerStatus(p.IP.String(), peer.BAD)
				continue
			}
			// defer c.Conn.Close()
			if msg, err := c.Read(); err == nil && msg.ID == message.MsgBitfield {
				c.Bitfield = msg.Payload
			}
			wg.Add(1)
			go func(peerIp net.IP) {
				defer c.Conn.Close()
				defer wg.Done()

				for !piecesQueue.IsEmpty() {
					var pieceIndex = heap.Pop(&piecesQueue).(uint)

					if !c.Bitfield.HasPiece(int(pieceIndex)) {
						fmt.Printf("%v does not have piece #%v\n", peerIp, pieceIndex)
						heap.Push(&piecesQueue, lib.NewPriorityItem(pieceIndex, max(int(endPiece-pieceIndex), 1)))
						continue
					}
					fmt.Printf("Piece #%d -> %v\n", pieceIndex, peerIp)
					pieceBuffer := make([]byte, t.PieceLength)

					err := c.DownloadPiece(pieceBuffer, pieceIndex, blockCount)
					if err != nil {
						fmt.Printf("Error downloading piece #%v: %v Dropping peer %v\n", pieceIndex, err, peerIp)
						heap.Push(&piecesQueue, lib.NewPriorityItem(pieceIndex, max(int(endPiece-pieceIndex), 1)))
						return
					}

					fmt.Printf("Piece #%v downloaded. bytes: %v\n", pieceIndex, len(pieceBuffer))

					// Compare sha1 hash against metadata checksum
					pieceHash := sha1.Sum(pieceBuffer[:])
					checkSumMatch := pieceHash == [20]byte(t.PieceHashes[pieceIndex])
					if !checkSumMatch {
						fmt.Printf("Checksum Fail! for piece #%v\n", pieceIndex)

						heap.Push(&piecesQueue, lib.NewPriorityItem(pieceIndex, max(int(endPiece-pieceIndex), 1)))
						continue
					}
					fmt.Printf("Matching Checksums for piece #%v!\n", pieceIndex)

					byteOffset := int64((pieceIndex - startPiece) * t.PieceLength)

					fmt.Printf("Writing piece #%v at byte offset %v\n", pieceIndex, byteOffset)
					fileLock.Lock()
					f.WriteAt(pieceBuffer[:], byteOffset)
					fileLock.Unlock()
				}
			}(p.IP)
		}
		fmt.Println("finished loop")
	}()

	// Start timer
	fmt.Println("starting timer")
	start := time.Now()
	wg.Wait()
	fmt.Printf("Downloaded %s in %s\n", t.Name, time.Since(start))
}

func FromMetadata(metadata []byte) (Torrent, error) {
	var t Torrent

	if err := bencode.Unmarshal(bytes.NewReader(metadata), &t); err != nil {
		fmt.Println("Error decoding metadata:", err)
		return t, err
	}

	return t, nil
}

func (t Torrent) WriteMetadataFile(dir string) error {
	filename := path.Join(dir, t.Name+".torrent")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	save := TorrentOut{
		Info_hash:     t.InfoHash,
		Announce_list: t.Trackers.String(),
		Name:          t.Name,
		Pieces:        t.PieceHashesString,
		PieceLength:   t.PieceLength,
	}
	if len(t.Files) > 0 {
		for _, file := range t.Files {
			save.Files = append(save.Files, FileOut(file))
		}
	} else {
		save.Length = t.Length
	}
	err = bencode.Marshal(f, save)
	return err
}

func (t *Torrent) String() string {
	return fmt.Sprintf("Torrent: %s\nFiles: \n%sInfoHash: %s\nPieceHashes: %v\nLength:%v\nPieceLength:%v", t.Name, t.Files, t.InfoHash, t.PieceHashes, t.Length, t.PieceLength)
}
