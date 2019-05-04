package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

	baseDir := flag.String("d", CWD(), "Use a custom directory to resolve relative paths from "+goup+".yml.")
	buildFile := flag.String("b", "./"+goup+".yaml", "Use a build file to load.")
	homeDir := flag.String("c", defaultHome, "Use this as the home directory, where "+GoUp+" holds toolchains, projects and workspaces.")
	logLevel := flag.Int("l", int(Error), "The LogLevel determines what is printed into the console. 0=Debug, 1=Info, 2=Warn, 3=Error")
	resourcesUrl := flag.String("r", defaultResourcesUrl, "XML which describes downloadable toolchains")

	showVersion := flag.Bool("v", false, "Shows the version")
	showHelp := flag.Bool("help", false, "Shows this help")

	flag.Parse()
	if *showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println(GoUp + " " + version)
		os.Exit(0)
	}

	a.BaseDir = Path(*baseDir)
	a.BuildFile = Path(*buildFile).Resolve(a.BaseDir)
	a.HomeDir = Path(*homeDir)
	a.LogLevel = LogLevel(*logLevel)
	a.ResourcesUrl = *resourcesUrl

	logger.Debug(Fields{"Name": GoUp, "Version": version, "GOARCH": runtime.GOARCH, "GOOS": runtime.GOOS})
	logger.Debug(Fields{"BaseDir": a.BaseDir, "BuildFile": a.BuildFile, "HomeDir": a.HomeDir, "LogLevel": a.LogLevel, "ResourcesUrl": a.ResourcesUrl})
}
