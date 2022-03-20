package lib

import (
	"fmt"
	"io"
	"net"
	"time"
)

const UDP_Timeout = time.Duration(15) * time.Second

func UDPRequest(address string, reader io.Reader) (res []byte, err error) {
	returnAddr, err := net.ResolveUDPAddr("udp", address)

	conn, err := net.DialUDP("udp", nil, returnAddr)
	if err != nil {
		fmt.Println("Error dialing UDP tracker:", err)
		return nil, err
	}
	defer conn.Close()

	doneChan := make(chan error, 1)
	bytesChan := make(chan []byte, 1)
	go func() {
		buffer := make([]byte, 1024)

		// Copy to the connection from the reader
		_, err := io.Copy(conn, reader)
		if err != nil {
			doneChan <- err
			return
		}

		deadline := time.Now().Add(UDP_Timeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			doneChan <- err
			return
		}

		nRead, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			doneChan <- err
			return
		}
		fmt.Printf("packet-received: bytes=%d from=%s\n", nRead, addr.String())
		bytesChan <- buffer
		doneChan <- nil
	}()

	select {
	case err = <-doneChan:
	case res = <-bytesChan:
	}
	return res, err
}
