// Torrent server

// Download a torrent file
// torrent magnet link:
// magnet:?xt=urn:btih:E7D80892BBCE0BDD761D38781DA480D9E64B1848&dn=David.Attenborough.A.Life.on.Our.Planet.2020.1080p.NF.WEBRip.DDP5.1.Atmos.x264-NTG&tr=http%3A%2F%2Ftracker.trackerfix.com%3A80%2Fannounce&tr=udp%3A%2F%2F9.rarbg.me%3A2810%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2930%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=http%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.zer0day.to%3A1337%2Fannounce

// full endpoint: /download/{formatted_magnet_link}
// Example: /download?xt=urn:btih:E7D80892BBCE0BDD761D38781DA480D9E64B1848&dn=David.Attenborough.A.Life.on.Our.Planet.2020.1080p.NF.WEBRip.DDP5.1.Atmos.x264-NTG&tr=http%3A%2F%2Ftracker.trackerfix.com%3A80%2Fannounce&tr=udp%3A%2F%2F9.rarbg.me%3A2810%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2930%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=http%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.zer0day.to%3A1337%2Fannounce
// Read the magnet link from the request;

package main

import (
	"fmt"
	"log"
	"net/http"

	torrent "torrent-pi/torrent"
)

const PORT int = 8080

func main() {
	http.HandleFunc("/download", download)
	fmt.Println("Listening on port:", PORT)
	log.Fatal(http.ListenAndServe(":"+fmt.Sprint(PORT), nil))
}

func download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Initiating download")

	torrent, err := torrent.NewTorrent(r.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}
	go torrent.Download()

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Torrent file downloading...")
}
