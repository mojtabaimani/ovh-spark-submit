package main

import (
	"fmt"
	"github.com/ncw/swift"
	"os"
	"path/filepath"
	"strings"
)

func UploadJar(address string, conn swift.Connection) {

	address = strings.TrimSpace(address)

	if !strings.HasPrefix(strings.ToLower(address), "swift://") {   // if address does not start with swift://

		jarname := filepath.Base(address)

		err := conn.ContainerCreate("jar", nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		content, err := os.Open(address)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer content.Close()

		fmt.Println("Uploading jar file to swift ... ")

		headers, err := conn.ObjectPut("jar", jarname, content, false, "", "application/zip", nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(headers)
		fmt.Println("Jar file uploaded to swift storage.")
	}
	wg.Done()

}
