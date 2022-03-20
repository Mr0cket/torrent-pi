package main

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
