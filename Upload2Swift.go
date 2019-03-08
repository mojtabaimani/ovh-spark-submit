package main

import (
	"fmt"
	"github.com/ncw/swift"
	"os"
	"path/filepath"
	"strings"
)

// if input file starts with swift:// copy it to the container/folder/object address
// otherwise it will upload it to swift in address container/folder/filename address
func Upload2Swift(address string, container string, folder string, conn swift.Connection) {

	address = strings.TrimSpace(address)

	if strings.HasPrefix(strings.ToLower(address), "swift://") { // if address starts with swift://
		inSwiftAddress :=address[8:] // remove "swift://" from beginning of the address
		parts := strings.Split(inSwiftAddress, "/")
		srcContainer := parts[0]
		srcObject:=parts[len(parts)-1]
		srcFolderObject := strings.TrimLeft(inSwiftAddress, srcContainer+"/")

		conn.ObjectCopy(srcContainer,srcFolderObject,container,folder+"/"+srcObject, nil)

		fmt.Println("\nObject "+srcObject+" was copied from swift://"+srcContainer+"/"+srcFolderObject+" to swift://"+
			container+"/"+folder+"/"+srcObject)
	} else {
		filename := filepath.Base(address)

		err := conn.ContainerCreate(container, nil)
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

		fmt.Println("Uploading input file to swift: ", address)

		_, err = conn.ObjectPut(container, folder+"/"+filename, content, false, "", "application/zip", nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("\nInput file uploaded to swift storage.", address)
	}
	wg.Done()

}
