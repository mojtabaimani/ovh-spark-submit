package main

import (
	"fmt"
	"github.com/ncw/swift"
	"os"
)

func Authenticate() swift.Connection {

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

	return conn

}
