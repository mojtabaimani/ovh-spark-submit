package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	flagclass := flag.String("class", "", "Your application's main class (for Java / Scala apps)")
	flagname := flag.String("name", "noName", "A name of your application")
	flagjars := flag.String("jars", "", `Comma-separated list of jars to include on the driver 
and executor classpaths`)
	flagpackages := flag.String("packages", "",
		`Comma-separated list of maven coordinates of jars to include 
on the driver and executor classpaths. Will search the local
maven repo, then maven central and any additional remote
repositories given by --repositories. The format for the
coordinates should be groupId:artifactId:version.`)
	flagexcludepackages := flag.String("exclude-packages", "",
		`Comma-separated list of groupId:artifactId, to exclude while
resolving the dependencies provided in --packages to avoid
dependency conflicts.`)
	flagrepositories := flag.String("repositories", "",
		`Comma-separated list of additional remote repositories to
search for the maven coordinates given with --packages.`)
	flagpyfiles := flag.String("py-files", "",
		`Comma-separated list of .zip, .egg, or .py files to place
on the PYTHONPATH for Python apps.`)
	flagfiles := flag.String("files", "",
		`Comma-separated list of files to be placed in the working
directory of each executor. File paths of these files
in executors can be accessed via SparkFiles.get(fileName).`)
	flagconf := flag.String("conf", "", "Arbitrary Spark configuration property.")
	flagpropertiesfile := flag.String("properties-file", "",
		`Path to a file from which to load extra properties. If not
specified, this will look for conf/spark-defaults.conf.`)
	flagdrivermemory := flag.String("driver-memory", "", "Memory for driver (e.g. 1000M, 2G) (Default: 1024M)")
	flagdriverjavaoptions := flag.String("driver-java-options", "", "Extra Java options to pass to the driver")
	flagdriverlibrarypath := flag.String("driver-library-path", "", "Extra library path entries to pass to the driver")
	flagdriverclasspath := flag.String("driver-class-path", "",
		`Extra class path entries to pass to the driver. Note that
jars added with --jars are automatically included in the
classpath.`)
	flagexecutormemory := flag.String("executor-memory", "", "Memory per executor (e.g. 1000M, 2G) (Default: 1G)")
	flagproxyuser := flag.String("proxy-user", "", "User to impersonate when submitting the application")
	flagverbose := flag.Bool("verbose", false, "Print additional debug output")
	flagversion := flag.String("version", "2.3", "Version of Spark")
	flagdrivercores := flag.Int("driver-cores", 1, "Number of cores used by the driver, only in cluster mode")
	flagsupervise := flag.Bool("supervise", false, "If given, restarts the driver on failure")
	flagtotalexecutorcores := flag.Int("total-executor-cores", 2, "Total cores for all executors")
	flagexecutorcores := flag.Int("executor-cores", 0,
		`Number of cores per executor. (Default: 1 in YARN mode
or all available cores on the worker in standalone mode)`)

	////////////////////////

	flagtoken := flag.String("token", "unknown", "Openstack token for authentication")
	///////////////////////

	//to avoid "declared and not used" error for all spark-submit commandline options.
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _ = flagclass, flagconf, flagdriverclasspath, flagdrivercores, flagdriverjavaoptions, flagdriverlibrarypath, flagdrivermemory, flagexcludepackages, flagexecutorcores, flagexecutormemory, flagfiles, flagjars, flagname, flagpackages, flagpropertiesfile, flagproxyuser, flagpyfiles, flagrepositories, flagsupervise, flagtotalexecutorcores, flagverbose, flagversion

	flag.Parse()

	if *flagtoken == "unknown" {
		fmt.Println("Please enter a token")
		panic(5)
	}

	jarpath := flag.Arg(0)
	jarname := filepath.Base(jarpath)
	_ = jarname

	fmt.Println("name:", *flagname)
	fmt.Println("Jar File:", jarpath)
	allArgs := fmt.Sprint(os.Args[1:])
	allArgs = strings.Replace(allArgs, "[", "", -1)
	allArgs = strings.Replace(allArgs, "]", "", -1)

	fmt.Println("all args:", allArgs)

	if _, err := os.Stat(jarpath); os.IsNotExist(err) {
		fmt.Println("Jar file does not exist.")
		panic(err)
	}

	file, err := os.Open(jarpath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	ServerAddress := "http://51.38.226.197:8090"

	req, err := http.NewRequest("POST", ServerAddress+"/upload?filename="+jarname, file)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))

	type Result struct {
		SessionID string `json:"sessionID"`
	}
	var result Result
	err2 := json.Unmarshal(body, &result)
	if err2 != nil {
		panic(err2)
	}

	fmt.Println("Session ID: ", result.SessionID)

	resp2, err := http.PostForm(ServerAddress+"/sparksubmit", url.Values{"commandline": {allArgs}, "sessionID": {result.SessionID}, "name": {*flagname}})
	if err != nil {
		panic(err)
	}
	body2, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body2))

	fmt.Println("Spark job submitted. You can see the output log of your Spark job by this link: " + ServerAddress + "/output/?sessionID=" + result.SessionID)

}
