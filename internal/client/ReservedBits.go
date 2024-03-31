package client

import "fmt"

type ReservedBits [8]byte

func (r *ReservedBits) Has(bit int) bool {
	effectiveBit := bit - 1
	byteIndex := effectiveBit / 8               // Automatically truncated because bit is an int
	effectiveBitIndex := 7 - (effectiveBit % 8) // convert to big Endian
	return r[byteIndex]&(1<<effectiveBitIndex) != 0
}

func (r *ReservedBits) String() string {
	// Bit representation of the bytes concatenated
	bitString := ""
	for _, b := range r {
		bitString += fmt.Sprintf("%08b", b)
	}
	return bitString
}
