package utils

import (
	"fmt"
	"os"
	"strings"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func HandleHTTPError(err error) bool {
	if err != nil {
		errorString := err.Error()

		// Parse string to determine error type
		if strings.Contains(errorString, "timeout") {
			fmt.Println("tracker timeout")
			return true
		}
		fmt.Println("Error:", errorString)
		os.Exit(1)
		return true
	}
	return false
}

// Splits a string into array of byte arrays
func SplitStringToBytes(str string, splitSize int) [][]byte {
	byteSubStrings := [][]byte{}
	if len(str)%splitSize > 0 {
		fmt.Printf("WARNING: string is not mutiple of split size %d\n", splitSize)
	}
	for i := 0; i < len(str); i += splitSize {
		end := i + splitSize
		if end > len(str) {
			end = len(str)
		}
		byteSubStrings = append(byteSubStrings, []byte(str[i:end]))
	}
	return byteSubStrings
}
