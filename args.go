package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Args contains the arguments which have been used to invoke GoUp
type Args struct {
	// The BaseDir is used to resolve paths in goup.yaml
	BaseDir Path

	// The goup.yml file to use
	BuildFile Path

	// HomeDir is the place where GoUp caches toolchains, projects and workspaces.
	HomeDir Path

	// The LogLevel determines what is printed into the console
	LogLevel LogLevel

	// ResourcesUrl is used to update the external resources list
	ResourcesUrl string

	// Targets contains the different build targets, e.g. gomobile/android or gomobile/ios
	Targets []string

	// ClearWorkspace does not reuse the workspace
	ClearWorkspace bool
}

// Evaluate reads all flags and parses them into the receiver.
// On failures or if help requested it may also print to the console
// and exit
func (a *Args) Evaluate() {
	defaultHome, err := os.UserHomeDir()
	if err != nil {
		defaultHome = "/"
	}
	defaultHome = filepath.Join(defaultHome, "."+goup)

	overriddenDefaultHome := os.Getenv("GOUP_HOME")
	if len(overriddenDefaultHome) > 0 {
		defaultHome = overriddenDefaultHome
	}

	baseDir := flag.String("dir", CWD(), "Use a custom directory to resolve relative paths from "+goup+".yml.")
	buildFile := flag.String("buildFile", "./"+goup+".yaml", "Use a build file to load.")
	homeDir := flag.String("home", defaultHome, "Use this as the home directory, where "+goUp+" holds toolchains, projects and workspaces.")
	logLevel := flag.Int("loglevel", int(Error), "The LogLevel determines what is printed into the console. 0=Debug, 1=Info, 2=Warn, 3=Error")
	resourcesUrl := flag.String("resources", defaultResourcesUrl, "XML which describes downloadable toolchains")
	targets := flag.String("targets", "all", "The targets to build, e.g. gomobile/android or gomobile/ios. Can be concated by :")

	showVersion := flag.Bool("version", false, "Shows the version")
	showHelp := flag.Bool("help", false, "Shows this help")
	doReset := flag.Bool("reset", false, "Performs a reset, delete the home directory and exits")
	doClean := flag.Bool("clean", false, "Removes the project workspace, but keeps toolchains.")

	flag.Parse()
	if *showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println(goUp + " " + version)
		os.Exit(0)
	}

	a.BaseDir = Path(*baseDir)
	a.BuildFile = Path(*buildFile).Resolve(a.BaseDir)
	a.HomeDir = Path(*homeDir)
	a.LogLevel = LogLevel(*logLevel)
	a.ResourcesUrl = *resourcesUrl
	a.Targets = strings.Split(*targets, ":")
	a.ClearWorkspace = *doClean

	logger = &defaultLogger{a.LogLevel}

	logger.Debug(Fields{"Name": goUp, "Version": version, "GOARCH": runtime.GOARCH, "GOOS": runtime.GOOS})
	logger.Debug(Fields{"BaseDir": a.BaseDir, "BuildFile": a.BuildFile, "HomeDir": a.HomeDir, "LogLevel": a.LogLevel, "ResourcesUrl": a.ResourcesUrl, "Targets": a.Targets})

	if *doReset {
		err := os.RemoveAll(a.HomeDir.String())
		must(err)
	}
}
