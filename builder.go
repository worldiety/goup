package main

import (

)
/*
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

	// BuildConfiguration is the configuration to use, as loaded from BuildFile
	BuildConfig *BuildConfiguration

	// HomeDir denotes the home directory of the user executing this process
	HomeDir Path

	// Env contains the environment variables, which may change over time and from step to step
	Env map[string]string

	// CWD is the current working dir and will change from step to step
	CWD Path

	// BuildCache collects information from all of our builds
	BuildCache *BuildCache
}

// IsBuildRequired tries to detect if we need to build again. Because gomobile/cgo compiles really slowly we want to
// avoid that in any case (e.g. 30s for hello world on a beefy machine) which takes a fraction of a second
// for go itself.
func (b *Builder) IsBuildRequired() bool {
	cacheFile := b.BuildDir.Child("build.cache")
	b.BuildCache = &BuildCache{}
	err := b.BuildCache.Load(cacheFile.String())
	if err != nil && b.Verbose {
		b.PP("failed to load the build cache file, could be normal: %v", err)
		return true
	}

	inHash := b.calculateInHash()
	outHash := b.calculateOutHash()

	if b.BuildCache.InHash != inHash || b.BuildCache.OutHash != outHash {
		if b.Verbose {
			b.PP("build cache indicates file changes")
		}
		return true
	}
	return false
}

func (b *Builder) UpdateBuildCache() {
	inHash := b.calculateInHash()
	outHash := b.calculateOutHash()

	b.BuildCache.InHash = inHash
	b.BuildCache.OutHash = outHash
	cacheFile := b.BuildDir.Child("build.cache")
	err := b.BuildCache.Save(cacheFile.String())
	if err != nil {
		fmt.Println(err)
	}
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

	b.BuildConfig = cfg
	b.BuildDir = b.HomeDir.Child(cfg.Project)
	b.GoPath = b.HomeDir.Child(cfg.Project).Child("workspace")


	if len(b.BuildConfig.Build.Android.Out) == 0 {
		b.BuildConfig.Build.Android.Out = Path("./" + b.BuildConfig.Project + ".aar")
	}

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

func (b *Builder) Gomobile() error {
	if b.Verbose {
		b.PP("\ninvoking gomobile...")
	}
	b.Chdir(b.GoPath)
	b.SetEnv("GO111MODULE", "off")

	if b.BuildConfig.Build.Android != nil {
		args := []string{"bind"}
		if b.Verbose {
			args = append(args, "-v")
		}


		outFile := b.BuildConfig.Build.Android.Out.Resolve(b.BaseDir)
		args = append(args, "-o", outFile.String())

		if len(b.BuildConfig.Build.Android.Package) > 0 {
			args = append(args, "-javapkg", b.BuildConfig.Build.Android.Package)
		}
		args = append(args, "-target=android")

		args = append(args, b.BuildConfig.Exports...)
		err := b.Run("bin/gomobile", args...)
		if err != nil {
			return err
		}

	}

	if b.BuildConfig.Build.IOS != nil {
		args := []string{"bind"}
		if b.Verbose {
			args = append(args, "-v")
		}

		if len(b.BuildConfig.Build.IOS.Out) == 0 {
			b.BuildConfig.Build.IOS.Out = Path("./" + b.BuildConfig.Project + ".framework")
		}
		outFile := b.BuildConfig.Build.IOS.Out.Resolve(b.BaseDir)
		args = append(args, "-o", outFile.String())

		if len(b.BuildConfig.Build.IOS.Prefix) > 0 {
			args = append(args, "-prefix", b.BuildConfig.Build.IOS.Prefix)
		}
		args = append(args, "-target=ios")

		args = append(args, b.BuildConfig.Exports...)
		err := b.Run("bin/gomobile", args...)
		if err != nil {
			return err
		}
	}
	return nil
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
func (b *Builder) CopyModulesToWorkspace() error {
	dependencies := make(map[string]VendoredModule)
	b.Chdir(b.GoPath)
	b.SetEnv("GO111MODULE", "on")
	for _, modPath := range b.BuildConfig.Imports {
		resolvedPath := modPath.Resolve(b.BaseDir)
		if b.Verbose {
			b.PP("processing module %s", resolvedPath)
		}
		modName, err := getModuleName(resolvedPath.Child("go.mod"))
		if err != nil {
			b.Failf("expected '%s' to have a go.mod file. This is not a go module: %v", resolvedPath, err)
		}
		if b.Verbose {
			b.PP("name is '%s'", modName)
		}

		// copy declared go modules into go path
		targetDir := b.GoPath.Child("src").Add(Path(modName))
		if b.Verbose {
			b.PP("removing to '%s'", targetDir)
		}
		err = os.RemoveAll(targetDir.String())
		if err != nil {
			b.Failf("failed to clear directory %s: %v", targetDir, err)
		}
		if b.Verbose {
			b.PP("copying to '%s'", targetDir)
		}

		err = CopyDir(resolvedPath.String(), targetDir.String())
		if err != nil {
			b.Failf("failed to copy directory %s: %v", targetDir, err)
		}

		// vendor module dependencies for each module
		b.Chdir(targetDir)
		err = b.Run("go", "mod", "vendor")
		if err != nil {
			b.Failf("failed to vendor module dependencies: %v", err)
		}

		modules, err := ParseModulesTxT(targetDir.Child("vendor").Child("modules.txt").String())
		if err != nil {
			b.Failf("failed to parse vendor module information: %v", err)
		}

		// collected and inspect all modules: upgrade to the largest declared version, causing potential semver conflict
		for _, mod := range modules {
			if b.Verbose {
				b.PP("dependency %s@%s", mod.ModuleName, mod.Version.String())
			}
			dep, ok := dependencies[mod.ModuleName]

			if !ok || mod.Version.IsNewer(dep.Version) {
				dependencies[mod.ModuleName] = mod
				dep = mod

				if b.Verbose {
					b.PP("module %s upgraded to %s", mod.ModuleName, mod.Version.String())
				}
			}
		}
	}

	// we collected all dependencies, now copy it into the workspace/gopath
	for _, dep := range dependencies {
		targetDir := b.GoPath.Child("src").Add(Path(dep.ModuleName))
		err := os.RemoveAll(targetDir.Parent().String())
		if err != nil {
			return fmt.Errorf("failed to remove module target directory: %v", err)
		}
		err = os.MkdirAll(targetDir.Parent().String(), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create module target directory: %v", err)
		}
		if b.Verbose {
			b.PP("moving dependency %s->%s", dep.Local.String(), targetDir.String())
		}
		err = os.Rename(dep.Local.String(), targetDir.String())
		if err != nil {
			return fmt.Errorf("failed to move: %s->%s: %v", dep.Local.String(), targetDir.String(), err)
		}
	}

	// clear all vendor directories in copied modules
	for _, modPath := range b.BuildConfig.Imports {
		resolvedPath := modPath.Resolve(b.BaseDir)
		modName, err := getModuleName(resolvedPath.Child("go.mod"))
		if err != nil {
			panic(err) // handled above already
		}
		targetDir := b.GoPath.Child("src").Add(Path(modName)).Add("vendor")

		if b.Verbose {
			b.PP("removing vendor folder from module: %s", targetDir)
		}
		err = os.RemoveAll(targetDir.String())
		if err != nil {
			return fmt.Errorf("failed to remove: %s: %v", targetDir, err)
		}

	}
	return nil
}

// StopWatch just prints a duration
func (b *Builder) StopWatch(t time.Time, msg string) {
	b.PP("%s: %s", msg, time.Now().Sub(t))
}
*/