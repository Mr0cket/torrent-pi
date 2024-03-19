package torrent

import "fmt"

type File struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

func (f File) String() string {
	return fmt.Sprintf("File: %s (%d)", f.Path[0], f.Length)
}

type Files []File

func (f Files) String() string {
	var s string
	for _, file := range f {
		s += fmt.Sprintf("%s\n", file)
	}
	return s
}
