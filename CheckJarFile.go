package main

import (
	"fmt"
	"github.com/ncw/swift"
	"os"
	"strings"
)

func CheckJarFile(address string, conn swift.Connection) {

	address = strings.TrimSpace(address)

	if strings.HasPrefix(strings.ToLower(address), "swift://") {

		address=address[8:len(address)] // remove swift:// from beginning of the address
		parts := strings.Split(address, "/")
		containerName := parts[0]
		objectName := strings.TrimLeft(address, containerName+"/")

		_, _, err := conn.Object(containerName, objectName)

		if err!=nil {
			fmt.Println(err)
			fmt.Println("Container:"+containerName)
			fmt.Println("Object:"+objectName)
			fmt.Println("Jar file didn't found in OpenStack Swift storage.")
			os.Exit(1)
		}

	} else {
		if _, err := os.Stat(address); os.IsNotExist(err) {
			fmt.Println("Jar file didn't found.")
			fmt.Println(address)
			os.Exit(1)
		}
	}
}
