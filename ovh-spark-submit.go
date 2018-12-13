package main

import (
	"flag"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {

	flagclass := flag.String("class", "", "Your application's main class (for Java / Scala apps)")
	flagname := flag.String("name", "NoName", "A name of your application")
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
	flagdrivermemory := flag.String("driver-memory", "",
		"Memory for driver (e.g. 1000M, 2G) (Default: 1024M)")
	flagdriverjavaoptions := flag.String("driver-java-options", "",
		"Extra Java options to pass to the driver")
	flagdriverlibrarypath := flag.String("driver-library-path", "",
		"Extra library path entries to pass to the driver")
	flagdriverclasspath := flag.String("driver-class-path", "",
		`Extra class path entries to pass to the driver. Note that
jars added with --jars are automatically included in the
classpath.`)
	flagexecutormemory := flag.String("executor-memory", "",
		"Memory per executor (e.g. 1000M, 2G) (Default: 1G)")
	flagproxyuser := flag.String("proxy-user", "",
		"User to impersonate when submitting the application")
	flagverbose := flag.Bool("verbose", false, "Print additional debug output")
	flagversion := flag.String("version", "2.4.0", "Version of Spark")
	flagdrivercores := flag.Int("driver-cores", 1,
		"Number of cores used by the driver, only in cluster mode")
	flagsupervise := flag.Bool("supervise", false, "If given, restarts the driver on failure")
	flagtotalexecutorcores := flag.Int("total-executor-cores", 2, "Total cores for all executors")
	flagexecutorcores := flag.Int("executor-cores", 0,
		`Number of cores per executor. (Default: 1 in YARN mode
or all available cores on the worker in standalone mode)`)
	flagkeepinfra := flag.Bool("keep-infra", false,
		"By using this flag, the spark cluster will not be deleted after finishing the job. ")
	flagdeployer := flag.String("deployer", "publicsequentialworkerjoin", "It selects deployer and infrastructure type.")
	flagnetwork := flag.String("network", "Ext-Net", "Private network name inside openstack project."+
		"If if is not mentioned, the cluster will be created in public network Ext-Net")

	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _ = flagclass, flagconf,
		flagdriverclasspath, flagdrivercores, flagdriverjavaoptions, flagdriverlibrarypath, flagdrivermemory,
		flagexcludepackages, flagexecutorcores, flagexecutormemory, flagfiles, flagjars, flagname, flagpackages,
		flagpropertiesfile, flagproxyuser, flagpyfiles, flagrepositories, flagsupervise, flagtotalexecutorcores,
		flagverbose, flagversion, flagkeepinfra, flagdeployer, flagnetwork

	flag.Parse()

	if *flagclass == "" {
		fmt.Println("Please enter the class name in --class option ")
		os.Exit(1)
	}

	CheckSparkVersion(*flagversion)

	jarpath := flag.Arg(0)

	fmt.Println("name:", *flagname)
	fmt.Println("Jar File:", jarpath)
	allArgs := fmt.Sprint(os.Args[1:])
	allArgs = strings.Replace(allArgs, "[", "", -1)
	allArgs = strings.Replace(allArgs, "]", "", -1)

	fmt.Println("all args:", allArgs)

	conn := Authenticate()

	CheckJarFile(jarpath, conn)

	wg.Add(1)

	go UploadJar(jarpath, conn)

	ServerAddress := "http://51.75.193.10:8090" //sparkalpha main server
	//ServerAddress := "http://145.239.28.145:8090"    //pre production test server

	id, err := uuid.NewV4()
	sessionID := id.String()
	fmt.Println("Session ID: " + sessionID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *flagdeployer == "vracfloatingip" && *flagnetwork == "Ext-Net" {
		*flagnetwork = "sparknetwork"
	}

	resp, err := http.PostForm(ServerAddress+"/sparksubmit", url.Values{"commandline": {allArgs},
		"sessionID": {sessionID}, "name": {*flagname}, "token": {conn.AuthToken}, "projectid": {conn.TenantId},
		"region": {conn.Region}, "deployer": {*flagdeployer}, "network": {*flagnetwork}})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(body))

	//fmt.Println("Spark job submitted. You can see the output log of your Spark job by this link: " +
	//	ServerAddress + "/output/?sessionID=" + sessionID)

	currentTime := time.Now()
	var home = os.Getenv("HOME")
	var logPath = home + "/SparkLogs/" + currentTime.Format("2006") + "/" + currentTime.Format("01") + "/" +
		currentTime.Format("02") + "/" + currentTime.Format("15-04-05-") + *flagname + "/"
	var logFullAddress = logPath + sessionID + ".log"
	os.MkdirAll(logPath, os.ModePerm)
	var offset = 0
	var output = ""
	for !strings.Contains(output, "Goodbye.") && !strings.Contains(output, "failed!!") {
		resp, err := http.Get(ServerAddress + "/output/?sessionID=" + sessionID + "&offset=" + strconv.Itoa(offset))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		offset += len(body)
		output = string(body)
		fmt.Print(output)

		f, err := os.OpenFile(logFullAddress, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if _, err := f.WriteString(output); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := f.Close(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		resp.Body.Close()
		time.Sleep(400 * time.Millisecond)
		//write output to the local log file

	}

	fmt.Println("Logs were saved in your openstack swift storage and also in your local machine at addresss: \n" +
		logFullAddress)

	wg.Wait()

}
