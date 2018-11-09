package main

import (
	"fmt"
	"github.com/ncw/swift"
	"os"
	"path/filepath"
	"strconv"
)

func UploadJar(address string) swift.Connection {

	fmt.Println("Uploading jar file to swift ... ")
	if _, err := os.Stat(address); os.IsNotExist(err) {
		fmt.Println("Jar file does not exist.")
		panic(err)
	}
	jarname := filepath.Base(address)

	authop, err := AuthOptionsFromEnv()

	conn := swift.Connection{
		Domain:         authop.DomainName,
		UserName:       authop.Username,
		ApiKey:         authop.Password,
		AuthUrl:        authop.IdentityEndpoint,
		Region:         authop.RegionName,
		TenantId:		authop.TenantID,
		AuthVersion:	authop.AuthVersion,
	}
	err = conn.Authenticate()
	if err != nil {
		fmt.Println(err)
		return nilCon
	}

	err = conn.ContainerCreate("jar", nil)
	if err != nil {
		fmt.Println(err)
		return nilCon
	}

	content, err := os.Open(address)
	if err != nil {
		fmt.Println(err)
		return nilCon
	}
	defer content.Close()

	headers, err := conn.ObjectPut("jar", jarname, content, false, "", "application/zip", nil)
	if err!=nil{
		fmt.Println(err)
		return nilCon
	}

	fmt.Println(headers)

	return conn
}

type AuthOptions struct {
	IdentityEndpoint string `json:"-"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	DomainName string `json:"name,omitempty"`
	TenantID   string `json:"tenantId,omitempty"`
	TenantName string `json:"tenantName,omitempty"`
	RegionName string `json:"-"`
	AuthVersion int `json:"-"`
}

func AuthOptionsFromEnv() (AuthOptions, error) {
	authURL := os.Getenv("OS_AUTH_URL")
	username := os.Getenv("OS_USERNAME")
	password := os.Getenv("OS_PASSWORD")
	tenantID := os.Getenv("OS_TENANT_ID")
	tenantName := os.Getenv("OS_TENANT_NAME")
	domainName := os.Getenv("OS_DOMAIN_NAME")
	regionName := os.Getenv("OS_REGION_NAME")
	authVersion:= os.Getenv("OS_IDENTITY_API_VERSION")

	// If OS_PROJECT_ID is set, overwrite tenantID with the value.
	if v := os.Getenv("OS_PROJECT_ID"); v != "" {
		tenantID = v
	}

	// If OS_PROJECT_NAME is set, overwrite tenantName with the value.
	if v := os.Getenv("OS_PROJECT_NAME"); v != "" {
		tenantName = v
	}

	if authURL == "" {
		// error
	}


	if password == "" {
		// error
	}

	if regionName == ""{
		// error
	}

	domainName="Default"

	version,_ := strconv.Atoi(authVersion)

	ao := AuthOptions{
		IdentityEndpoint:            authURL,
		Username:                    username,
		Password:                    password,
		TenantID:                    tenantID,
		TenantName:                  tenantName,
		DomainName:                  domainName,
		RegionName:					 regionName,
		AuthVersion: 				 version,
	}

	return ao, nil
}
var nilOptions = AuthOptions{}
var nilCon = swift.Connection{}
