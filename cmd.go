package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const version = "0.0.1"

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
	builder := &Builder{}
	builder.Init()
	builder.EnsureGoMobile()
	builder.CopyModulesToWorkspace()
}
