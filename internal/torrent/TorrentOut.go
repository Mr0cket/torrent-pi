package torrent

type FileOut struct {
	Length int      "length"
	Path   []string "path"
}

type TorrentOut struct {
	Info_hash     [20]byte "info_hash"
	Name          string   "name"
	Announce_list []string "announce_list"

	// pieces maps to a string whose length is a multiple of 20.
	// It is to be subdivided into strings of length 20, each of which is the SHA1 hash of the piece at the corresponding index.
	Pieces      string "pieces"
	PieceLength uint   "piece length"
	Length      uint   "length"

	// For the purposes of the other keys, the multi-file case is treated as only having a single file
	// by concatenating the files in the order they appear in the files list.
	Files []FileOut "files"
}
