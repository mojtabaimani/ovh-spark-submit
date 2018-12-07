package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

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
	if strings.ToLower(regionName) == "sbg3" {
		fmt.Println("Region SBG3 is not supported. Please change the region in OS_REGION_NAME variable.") //TODO: it should be changed if the PCI updated the openstack of SBG3 region.
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
