package main

import (
	"fmt"
	"github.com/ncw/swift"
	"os"
	"path/filepath"
	"strconv"
)

func UploadJar(address string) swift.Connection {

	if _, err := os.Stat(address); os.IsNotExist(err) {
		fmt.Println("Jar file does not exist.")
		os.Exit(1)
	}

	jarname := filepath.Base(address)

	authop, err := AuthOptionsFromEnv()

	conn := swift.Connection{
		Domain:      authop.DomainName,
		UserName:    authop.Username,
		ApiKey:      authop.Password,
		AuthUrl:     authop.IdentityEndpoint,
		Region:      authop.RegionName,
		TenantId:    authop.TenantID,
		AuthVersion: authop.AuthVersion,
	}
	err = conn.Authenticate()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = conn.ContainerCreate("jar", nil)
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

	return conn
}

type AuthOptions struct {
	IdentityEndpoint string `json:"-"`
	Username         string `json:"username,omitempty"`
	Password         string `json:"password,omitempty"`
	DomainName       string `json:"name,omitempty"`
	TenantID         string `json:"tenantId,omitempty"`
	TenantName       string `json:"tenantName,omitempty"`
	RegionName       string `json:"-"`
	AuthVersion      int    `json:"-"`
}

func AuthOptionsFromEnv() (AuthOptions, error) {
	authURL := os.Getenv("OS_AUTH_URL")
	username := os.Getenv("OS_USERNAME")
	password := os.Getenv("OS_PASSWORD")
	tenantID := os.Getenv("OS_TENANT_ID")
	tenantName := os.Getenv("OS_TENANT_NAME")
	domainName := os.Getenv("OS_DOMAIN_NAME")
	regionName := os.Getenv("OS_REGION_NAME")
	authVersion := os.Getenv("OS_IDENTITY_API_VERSION")

	// If OS_PROJECT_ID is set, overwrite tenantID with the value.
	if v := os.Getenv("OS_PROJECT_ID"); v != "" {
		tenantID = v
	}

	// If OS_PROJECT_NAME is set, overwrite tenantName with the value.
	if v := os.Getenv("OS_PROJECT_NAME"); v != "" {
		tenantName = v
	}

	if authURL == "" {
		fmt.Println("Environment variable OS_AUTH_URL is missing.")
		os.Exit(1)
	}

	if password == "" {
		fmt.Println("Environment variable OS_PASSWORD is missing.")
		os.Exit(1)
	}
	if username == "" {
		fmt.Println("Environment variable OS_USERNAME is missing.")
		os.Exit(1)
	}

	if regionName == "" {
		fmt.Println("Environment variable OS_REGION_NAME is missing.")
		os.Exit(1)
	}

	if tenantID == "" {
		fmt.Println("Environment variable OS_TENANT_ID or OS_PROJECT_ID is missing.")
		os.Exit(1)
	}

	domainName = "Default"

	version, _ := strconv.Atoi(authVersion)

	ao := AuthOptions{
		IdentityEndpoint: authURL,
		Username:         username,
		Password:         password,
		TenantID:         tenantID,
		TenantName:       tenantName,
		DomainName:       domainName,
		RegionName:       regionName,
		AuthVersion:      version,
	}

	return ao, nil
}

var nilOptions = AuthOptions{}
var nilCon = swift.Connection{}
