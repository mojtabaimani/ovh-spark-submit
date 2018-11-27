package main

import (
	"fmt"
	"os"
)

func CheckSparkVersion(version string) {
	sparkversions := []string {"1.6.3", "2.0.2" , "2.1.0" , "2.1.1" , "2.1.2" , "2.1.3" , "2.2.0" , "2.2.1" , "2.2.2" , "2.3.0" , "2.3.1" , "2.3.2" , "2.4.0" }

	exist, _ := inArray(version, sparkversions)
	if exist == false {
		fmt.Println("Wrong Spark version "+version + " , Possible Spark versions are: ")
		fmt.Println(sparkversions)
		os.Exit(1)
	}
	
}
