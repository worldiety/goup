package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const version = "0.0.2"

func evaluateFlags() (workingDir Path, buildFile Path, homeDir Path, verbose bool) {
	defaultHome, err := os.UserHomeDir()
	if err != nil {
		defaultHome = "/"
	}
	defaultHome = filepath.Join(defaultHome, ".gomobilebuilder")

	tdir := flag.String("dir", CWD(), "Sets the directory which is used to resolve relative paths. Default is the current working directory")
	tbuildFile := flag.String("buildFile", "./gomobile.build", "Sets the build file to load. Default is gomobile.build in the working directory")
	tverbose := flag.Bool("verbose", false, "Sets the verbose flag")
	tversion := flag.Bool("version", false, "Shows the version")
	thelp := flag.Bool("help", false, "Shows this help")
	thome := flag.String("home", defaultHome, "Sets the home directory, usually the ~ of the current user")

	flag.Parse()
	if *thelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *tversion {
		fmt.Println("gomobilebuilder " + version)
		os.Exit(0)
	}

	return Path(*tdir), Path(*tbuildFile).Resolve(Path(*tdir)), Path(*thome), *tverbose
}

func main() {
	total := time.Now()
	start := time.Now()

	builder := &Builder{}
	builder.Init()

	if !builder.IsBuildRequired() {
		builder.PP("everything is up to date, nothing to do")
		builder.StopWatch(total, "total")
		os.Exit(0)
	}


	builder.EnsureGoMobile()
	builder.StopWatch(start, "preparation")

	start = time.Now()
	err := builder.CopyModulesToWorkspace()
	if err != nil {
		fmt.Println("failed to prepare modules in workspace:", err)
		os.Exit(-1)
	}
	builder.StopWatch(start, "workspace setup")

	start = time.Now()
	err = builder.Gomobile()
	if err != nil {
		fmt.Println("failed to compile with gomobile:", err)
		os.Exit(-1)
	}
	builder.UpdateBuildCache()

	builder.StopWatch(start, "gomobile")
	builder.StopWatch(total, "total")

}