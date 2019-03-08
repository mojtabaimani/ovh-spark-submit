package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ncw/swift"
)

func CheckInputFile(address string, conn swift.Connection) {

	if address == "" {
		fmt.Println("Main application file is missing.")
		os.Exit(1)
	}

	address = strings.TrimSpace(address)

	if strings.HasPrefix(strings.ToLower(address), "swift://") {

		address = address[8:] // remove "swift://" from beginning of the address
		parts := strings.Split(address, "/")
		containerName := parts[0]
		objectName := strings.TrimLeft(address, containerName+"/")

		_, _, err := conn.Object(containerName, objectName)

		if err != nil {
			fmt.Println(err)
			fmt.Println("Container:" + containerName)
			fmt.Println("Object:" + objectName)
			fmt.Println("Input file didn't found in OpenStack Swift storage: ", address)
			os.Exit(1)
		}

	} else {
		if _, err := os.Stat(address); os.IsNotExist(err) {
			fmt.Println("Input file didn't found on local machine: ", address)
			os.Exit(1)
		}
	}
	wg.Done()
}
