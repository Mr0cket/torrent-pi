package client

import (
	"fmt"
	"strconv"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestReservedBitsHas(t *testing.T) {
	bitString := "0000000000000000000000000000000000000000000110000000000000000101"
	inputBits := []int{44, 63, 64, 35}
	expected := []bool{true, false, true, false}
	testBitArray := ReservedBits{}
	fmt.Println("bitString len", len(bitString)/8)
	for i := 0; i < len(bitString)/8; i++ {
		b, _ := strconv.ParseInt(bitString[i*8:(i+1)*8], 2, 8)
		testBitArray[i] = byte(b)
	}

	for i, char := range bitString {
		if char == rune("1"[0]) {
			fmt.Println("bit", i+1, "set")
		}
	}
	fmt.Println("test bits:", testBitArray.String())
	for index, bit := range inputBits {
		output := testBitArray.Has(bit)
		fmt.Printf("Has bit %v: %v\n", bit, output)
		if output != expected[index] {
			t.Fatalf(`ReservedBits.Has(%v) failed`, bit)
		}
	}
}
