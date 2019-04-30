package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Builder struct {
	// BaseDir is used to resolve relative paths
	BaseDir Path

	// BuildDir is the root of our build folder, somewhere in ~/.gomobilebuilder/<project>
	BuildDir Path

	// GoPath is the workspace directory of go, somewhere in ~./gomobilebuilder/<project>/workspace
	GoPath Path

	// Verbose is the flag to print annoying information
	Verbose bool

	// BuildFile is the file to load the config from
	BuildFile Path

	// Build is the configuration to use, as loaded from BuildFile
	Build *Build

	// HomeDir denotes the home directory of the user executing this process
	HomeDir Path

	// Env contains the environment variables, which may change over time and from step to step
	Env map[string]string

	// CWD is the current working dir and will change from step to step
	CWD Path
}

// Failf prints to the console and terminates the process
func (b *Builder) Failf(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
	if b.Verbose {
		b.PP("fatal error, cannot continue build")
	}
	os.Exit(-1)
}

// PP pretty prints a line into the console
func (b *Builder) PP(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}

// Chdir changes the CWD
func (b *Builder) Chdir(path Path) {
	b.CWD = path
	if b.Verbose {
		b.PP("cd %s", path)
	}
}

// SetEnv set a key/value environment variable
func (b *Builder) SetEnv(key string, val string) {
	b.Env[key] = val
	if b.Verbose {
		b.PP("export %s=%s", key, val)
	}
}

// Init loads and ensures all paths and config
func (b *Builder) Init() {
	workingDir, buildFile, home, verbose := evaluateFlags()

	b.BaseDir = workingDir
	b.Verbose = verbose
	b.BuildFile = buildFile
	b.HomeDir = home

	if b.Verbose {
		b.PP("verbose: on")
		b.PP("home directory: %s", b.HomeDir)
		b.PP("base directory: %s", b.BaseDir)
		b.PP("build file: %s", b.BuildFile)
	}

	mkdirs(b.BaseDir)
	mkdirs(b.HomeDir)

	cfg, err := LoadBuildFile(b.BuildFile)
	if err != nil {
		b.Failf("%v", err)
	}

	b.Build = cfg
	b.BuildDir = b.HomeDir.Child(cfg.Project)
	b.GoPath = b.HomeDir.Child(cfg.Project).Child("workspace")

	if b.Verbose {
		b.PP("loaded build configuration: %s", cfg.String())
		b.PP("build directory: %s", b.BuildDir)
		b.PP("GOPATH: %s", b.GoPath)
	}

	mkdirs(b.BuildDir)
	mkdirs(b.GoPath)

	b.Env = make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		b.SetEnv(pair[0], pair[1])
	}

	b.SetEnv("GOPATH", b.GoPath.String())

	if b.Verbose {
		b.PP("init done\n")
	}
}

func (b *Builder) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	for k, v := range b.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
		if b.Verbose {
			b.PP("%s=%s", k, v)
		}
	}
	if b.Verbose {
		s := name + " "
		for _, a := range args {
			s += a + " "
		}
		b.PP("%s", s)
	}
	cmd.Dir = b.CWD.String()
	stdoutStderr, err := cmd.CombinedOutput()
	b.PP("%s", stdoutStderr)
	if b.Verbose {
		b.PP("exec done\n")
	}

	return err
}

// EnsureGoMobile tries to execute gomobile
func (b *Builder) EnsureGoMobile() {
	b.Chdir(b.GoPath)
	b.SetEnv("GO111MODULE", "off")
	err := b.Run("bin/gomobile", "version")
	if err != nil {
		b.PP("%v", err)
		b.PP("installing gomobile")
		must(b.Run("go", "get", "-u", "golang.org/x/mobile/cmd/gomobile"))
		b.PP("installation complete")
		must(b.Run("bin/gomobile", "version"))

		must(b.Run("bin/gomobile", "init"))
	}
}

// CopyModulesToWorkspace does what the names says
func (b *Builder) CopyModulesToWorkspace() {
	b.Chdir(b.GoPath)
	b.SetEnv("GO111MODULE", "on")
	for _, modPath := range b.Build.Imports {
		resolvedPath := modPath.Resolve(b.BaseDir)
		if b.Verbose {
			b.PP("processing module %s", resolvedPath)
		}
		modName, err := getModuleName(resolvedPath.Child("go.mod"))
		if err != nil {
			b.Failf("expected '%s' to have a go.mod file. This is not a go module: %v", resolvedPath, err)
		}
		if b.Verbose {
			b.PP("  name is '%s'", modName)
		}

		// copy declared go modules into go path
		targetDir := b.GoPath.Child("src").Add(Path(modName))
		if b.Verbose {
			b.PP("  removing to '%s'", targetDir)
		}
		err = os.RemoveAll(targetDir.String())
		if err != nil {
			b.Failf("failed to clear directory %s: %v", targetDir, err)
		}
		if b.Verbose {
			b.PP("  copying to '%s'", targetDir)
		}

		err = CopyDir(resolvedPath.String(), targetDir.String())
		if err != nil {
			b.Failf("failed to copy directory %s: %v", targetDir, err)
		}

		// vendor module dependencies
		b.Chdir(targetDir)
		err = b.Run("go", "mod", "vendor")
		if err != nil {
			b.Failf("failed to vendor module dependencies: %v", err)
		}

	}
}

