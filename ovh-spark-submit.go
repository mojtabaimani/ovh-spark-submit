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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ncw/swift"
)

var wg sync.WaitGroup

func main() {

	flag.Usage = usage

	flagclass := flag.String("class", "",
		"Your application's main class (for Java / Scala apps)")
	flagname := flag.String("name", "NoName",
		"A name of your application")
	flag.Var(&inputJars, "jars",
		"Comma-separated list of jars to include on the driver \n"+
			"and executor classpaths")
	flagpackages := flag.String("packages", "",
		"Comma-separated list of maven coordinates of jars to include\n"+
			"on the driver and executor classpaths. Will search the local\n"+
			"maven repo, then maven central and any additional remote\n"+
			"repositories given by --repositories. The format for the\n"+
			"coordinates should be groupId:artifactId:version.\n")
	flagexcludepackages := flag.String("exclude-packages", "",
		"Comma-separated list of groupId:artifactId, to exclude while\n"+
			"resolving the dependencies provided in --packages to avoid\n"+
			"dependency conflicts.")
	flagrepositories := flag.String("repositories", "",
		"Comma-separated list of additional remote repositories to\n"+
			"search for the maven coordinates given with --packages.")
	flag.Var(&inputPys, "py-files",
		"Comma-separated list of .zip, .egg, or .py files to place\n"+
			"on the PYTHONPATH for Python apps.")
	flag.Var(&inputFiles, "files",
		"Comma-separated list of files to be placed in the working\n"+
			"directory of each executor. File paths of these files\n"+
			"in executors can be accessed via SparkFiles.get(fileName).")
	flag.Var(&sparkConf, "conf", "Arbitrary Spark configuration property.")
	flagpropertiesfile := flag.String("properties-file", "",
		"Path to a file from which to load extra properties. If not\n"+
			"specified, this will look for conf/spark-defaults.conf.")
	flagdrivermemory := flag.String("driver-memory", "",
		"Memory for driver (e.g. 1000M, 2G) (Default: 1024M)")
	flagdriverjavaoptions := flag.String("driver-java-options", "",
		"Extra Java options to pass to the driver")
	flagdriverlibrarypath := flag.String("driver-library-path", "",
		"Extra library path entries to pass to the driver")
	flagdriverclasspath := flag.String("driver-class-path", "",
		"Extra class path entries to pass to the driver. Note that\n"+
			"jars added with --jars are automatically included in the\n"+
			"classpath.")
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
		"Number of cores per executor. (Default: 1 in YARN mode\n"+
			"or all available cores on the worker in standalone mode)")
	flagkeepinfra := flag.Bool("keep-infra", false,
		"By using this flag, the spark cluster will not be deleted after finishing the job. ")
	flagnetworkname := flag.String("network-name", "sparknetwork", "Network name inside openstack project.")
	flagsubnetrange := flag.String("subnet-range", "192.168.18.0/24", "If the selected network is a private network, \n"+
		"the spark cluster will be created in this subnet range.")

	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _ = flagclass, sparkConf,
		flagdriverclasspath, flagdrivercores, flagdriverjavaoptions, flagdriverlibrarypath, flagdrivermemory,
		flagexcludepackages, flagexecutorcores, flagexecutormemory, inputPys, inputJars, flagname, flagpackages,
		flagpropertiesfile, flagproxyuser, inputFiles, flagrepositories, flagsupervise, flagtotalexecutorcores,
		flagverbose, flagversion, flagkeepinfra, flagnetworkname, flagsubnetrange

	flag.Parse()

	CheckSparkVersion(*flagversion)

	mainApp := flag.Arg(0)
	allArgs := strings.Join(os.Args[1:], " ")

	fmt.Println("name:", *flagname)
	fmt.Println("All args:", allArgs)
	fmt.Println("conf: ", sparkConf)
	fmt.Println("file:", strings.Join(inputFiles[:], ","))
	fmt.Println("jars:", strings.Join(inputJars[:], ","))
	fmt.Println("py-files:", strings.Join(inputPys[:], ","))
	fmt.Println("Main application file:", mainApp)

	conn := Authenticate()

	//check if all input files exist (check in parallel)
	if *flagpropertiesfile != "" {
		wg.Add(1)
		CheckInputFile(*flagpropertiesfile, conn)
	}
	wg.Add(1)
	go CheckInputFile(mainApp, conn)
	Map(inputFiles, conn, CheckInputFile)
	Map(inputJars, conn, CheckInputFile)
	Map(inputPys, conn, CheckInputFile)
	wg.Wait()

	ServerAddress := "http://51.38.224.115:8090" //sparkalpha server

	var deployer = "vrackfloatingip"
	if *flagnetworkname == "Ext-Net" {
		deployer = "public"
	}

	//just list of the files is enough, because all of them will be in /home/ubuntu directory in spark master node
	deployerArgs := " --name " + *flagname + " --token " + conn.AuthToken +
		" --project-id " + conn.TenantId + " --region " + conn.Region + " --network-name " + *flagnetworkname +
		" --subnet-range " + *flagsubnetrange + " " + allArgs //allArgs should be at the end because jar file and arguments should be at the end.
	if len(inputFiles) > 0 {
		deployerArgs += " --all-input-files " + AllFiles(inputFiles)
	}
	if len(inputJars) > 0 {
		deployerArgs += " --all-input-jars " + AllFiles(inputJars)
	}
	if len(inputPys) > 0 {
		deployerArgs += " --all-input-pys " + AllFiles(inputPys)
	}

	resp, err := http.PostForm(ServerAddress+"/", url.Values{"deployerArgs": {deployerArgs},
		"deployer": {deployer}, "name": {*flagname}})
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

	var cluster ClusterDocument
	json.Unmarshal(body, &cluster)

	if cluster.Status != "Successful" {
		fmt.Println("Cluster creation failed!")
		fmt.Println(cluster)
		os.Exit(1)
	}

	fullSwiftAddress := cluster.SwiftContainer + "/" + cluster.SwiftFolder
	fmt.Println("Spark files and logs address in Swift storage is: ", fullSwiftAddress)
	fmt.Println("Your cluster ID is: " + cluster.ClusterId)

	if *flagpropertiesfile != "" {
		wg.Add(1)
		Upload2Swift(*flagpropertiesfile, cluster.SwiftContainer, cluster.SwiftFolder, conn)
	}
	wg.Add(1)
	go Upload2Swift(mainApp, cluster.SwiftContainer, cluster.SwiftFolder, conn)
	Map2(inputFiles, cluster.SwiftContainer, cluster.SwiftFolder, conn, Upload2Swift)
	Map2(inputJars, cluster.SwiftContainer, cluster.SwiftFolder, conn, Upload2Swift)
	Map2(inputPys, cluster.SwiftContainer, cluster.SwiftFolder, conn, Upload2Swift)

	fmt.Println("Waiting for all upload operations...")
	wg.Wait() //waiting for all uploads to complete.

	var home = os.Getenv("HOME")
	var logPath = home + "/SparkLogs/" + cluster.SwiftFolder + "/"
	var logFullAddress = logPath + cluster.ClusterId + ".log"
	os.MkdirAll(logPath, os.ModePerm)
	fmt.Println("Log file created at " + logFullAddress)
	var offset = 0
	var output = ""
	for !strings.Contains(output, "Goodbye.") && !strings.Contains(output, "failed!!") {
		resp, err := http.Get(ServerAddress + "/" + cluster.ClusterId + "/logs?offset=" + strconv.Itoa(offset))
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
	}

	fmt.Println("\nLogs were saved in your openstack swift storage and also in your local machine at addresss: \n" +
		logFullAddress)

}

type inputFileList []string

func (i *inputFileList) String() string {
	return "hello"
}

func (i *inputFileList) Set(value string) error {
	var tmp = strings.Split(value, ",")
	for _, element := range tmp {
		*i = append(*i, element)
	}

	return nil
}

var inputJars, inputPys, inputFiles, sparkConf inputFileList

func usage() { //TODO: to be completed.
	fmt.Println("Usage:\n" +
		"ovh-spark-submit <option> \n" +
		"	options: \n" +
		"" +
		"--jar <your jar file> " +
		"")

}

func Map(array []string, conn swift.Connection, f func(string, swift.Connection)) {
	for _, v := range array {
		wg.Add(1)
		go f(v, conn)
	}
}
func Map2(array []string, container string, folder string, conn swift.Connection, f func(string, string, string, swift.Connection)) {
	for _, v := range array {
		wg.Add(1)
		go f(v, container, folder, conn)
	}
}
func AllFiles(array []string) string {
	var allFiles string
	for _, v := range array {
		allFiles += filepath.Base(v) + ","
	}
	allFiles = strings.TrimRight(allFiles, ",")
	return allFiles
}

type ClusterDocument struct {
	Status         string `json:"status"`
	ClusterId      string `json:"clusterId"`
	SwiftContainer string `json:"swiftContainer"`
	SwiftFolder    string `json:"swiftFolder"`
}
