package torrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"torrent-pi/internal/client"
	"torrent-pi/internal/constants"
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
type Torrent struct {
	PeerID   [20]byte   `bencode:"-"`
	InfoHash [20]byte   `bencode:"info_hash"`
	Name     string     `bencode:"name"`
	Trackers []*url.URL `bencode:"announce_list"`

	// pieces maps to a string whose length is a multiple of 20.
	// It is to be subdivided into strings of length 20, each of which is the SHA1 hash of the piece at the corresponding index.
	PieceHashesString string `bencode:"pieces"`
	PieceHashes       [][]byte
	PieceLength       uint `bencode:"piece length"`
	Length            uint `bencode:"length"`

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
		Name:        name,
		Trackers:    Trackers,
		PeerManager: peer.NewPeerManager(infoHash, []byte(constants.PEER_ID), Trackers),
	}
	copy(t.PeerID[:], []byte(constants.PEER_ID))
	copy(t.InfoHash[:], infoHash)

	// TODO Retrieve torrent metadata from the "swarm"... http://www.bittorrent.org/beps/bep_0009.html

	// Fetch peers from first tracker to respond
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

	// videoFile := string

	fmt.Println("##### Files #####")
	// Download only .mp4 files
	var fileToDownload File
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
	// startPiece := startByte / t.PieceLength
	// endPiece := (startByte + uint(fileToDownload.Length)) / t.PieceLength
	startPiece := uint(0)
	endPiece := uint(fileToDownload.Length) / t.PieceLength

	fmt.Printf("startPiece: %d, endPiece: %d\n", startPiece, endPiece)
	connections := make([]client.Client, 0)
	piecesQueue := make(chan uint, (endPiece-startPiece)+5)
	for i := startPiece; i <= endPiece; i++ {
		piecesQueue <- uint(i)
	}

	max_connections := 10
	wg := sync.WaitGroup{}
	wg.Add(max_connections)
	fileLock := sync.Mutex{}
	f, err := os.OpenFile(string(fileToDownload.Path[0]), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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
			msg, err := message.Read(c.Conn)
			fmt.Println("Pre-request message:", msg.TypeString())
			peerBitfield := message.ParseBitfield(msg)

			go func(peerIp net.IP) {
				defer wg.Done()
				// While there is items in piecesQueue,
				errorCount := 0
				for {
					select {
					case pieceIndex := <-piecesQueue:
						// Download each piece in 16 kb blocks
						blocks := t.PieceLength / constants.BLOCK_SIZE
						if !peerBitfield.HasPiece(int(pieceIndex)) {
							fmt.Printf("%v does not have piece #%v\n", peerIp, pieceIndex)
							piecesQueue <- pieceIndex
							continue
						}

						if pieceIndex == 420 {
							fmt.Println("🔥🔥🔥🔥🔥🔥")
						}
						fmt.Printf("Piece #%d -> %v\n", pieceIndex, peerIp)
						if pieceIndex == 420 {
							fmt.Println("🔥🔥🔥🔥🔥🔥")
						}

						pieceBuffer := make([]byte, t.PieceLength)
						for blockIndex := range make([]int, blocks) {
							var msg = &message.Message{}
							startByte := uint(blockIndex) * constants.BLOCK_SIZE

							for {
								fmt.Printf("Piece #%d block #%d -> %v\n", pieceIndex, blockIndex, peerIp)

								// Throw away all messages until we get a piece
								c.SendRequest(pieceIndex, startByte, constants.BLOCK_SIZE)
								for msg.ID != message.MsgPiece {
									msg, err = message.Read(c.Conn)
									if err != nil {
										msg = &message.Message{}
										errorCount++
										if errorCount > 3 {
											piecesQueue <- pieceIndex
											fmt.Printf("Dropping peer %v. Error limit reached\n", peerIp)
											// t.PeerManager.SetPeerStatus(p.IP.String(), peer.BAD)
											return
										}
										if err.Error() == "EOF" {
											fmt.Printf("Recieved EOF error. Re-requesting piece #%v block #%d\n", pieceIndex, blockIndex)
											break
										} else {
											fmt.Println("Error: ", err)
										}
									}
								}
								if msg.ID == message.MsgPiece {
									_, err := message.ParsePiece(int(pieceIndex), pieceBuffer, msg)
									if err != nil {
										fmt.Println("Buffer error", err)
									}
									break
								}
							}
						}

						fmt.Printf("Piece #%v downloaded. bytes: %v\n", pieceIndex, len(pieceBuffer))
						// Compare sha1 hash against metadata checksum
						pieceHash := sha1.Sum(pieceBuffer[:])
						checkSumMatch := pieceHash == [20]byte(t.PieceHashes[pieceIndex])
						if !checkSumMatch {
							fmt.Printf("Checksum Fail! for piece #%v\n", pieceIndex)

							// drop piece
							piecesQueue <- pieceIndex
							continue
						}
						fmt.Printf("Matching Checksums for piece #%v!\n", pieceIndex)

						byteOffset := int64((pieceIndex - startPiece) * t.PieceLength)
						fmt.Printf("Writing piece #%v at byte offset %v\n", pieceIndex, byteOffset)

						fileLock.Lock()
						f.WriteAt(pieceBuffer[:], byteOffset)
						fileLock.Unlock()

					case <-time.After(100 * time.Millisecond):
						fmt.Printf("Peer %v: piecesQueue empty\n", peerIp)
						return
					}
				}
			}(p.IP)
		}
		fmt.Println("finished loop")
	}()

	// Start timer
	fmt.Println("starting timer")
	start := time.Now()
	wg.Wait()
	elapsedTime := time.Since(start)
	fmt.Printf("Execution Time: %s\n", elapsedTime)

	// NiceToHave: Downloading blocks in ascending order to enable streaming

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
	// filename := path.Join(dir, t.Name+".torrent")
	// f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	f := new(bytes.Buffer)
	err := bencode.Marshal(f, t)
	fmt.Println(f.String())
	return err
}

func (t *Torrent) String() string {
	return fmt.Sprintf("Torrent: %s\nFiles: \n%sInfoHash: %s\nPieceHashes: %v\nLength:%v\nPieceLength:%v", t.Name, t.Files, t.InfoHash, t.PieceHashes, t.Length, t.PieceLength)
}
