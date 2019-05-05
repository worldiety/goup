package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GoUp contains the actual state of the GoUp program
type GoUp struct {
	// The program arguments
	args *Args
	// The parsed config
	config *GoUpConfiguration

	// the buildDir is the folder where we collect everything for this project
	buildDir Path

	// resources contains a list of valid and known resources, this is always an incomplete list, but
	// users may also choose their own list
	resources *Resources

	// env contains the environment variables, which may change over time and from step to step.
	// Initially this contains exact those variables with which this program has been launched.
	env map[string]string

	// cwd is the current working directory and used to launch external programs
	cwd Path

	// artifactCache contains information about the last build and is used to avoid unnecessary builds
	artifactCache *ArtifactCache
}

// NewGoupBuilder creates a new GoUp builder
func NewGoup(args *Args) (*GoUp, error) {
	gp := &GoUp{}
	gp.args = args
	gp.config = &GoUpConfiguration{}
	err := gp.config.Load(gp.args.BuildFile)
	if err != nil {
		return nil, err
	}

	logger.Debug(Fields{"buildFile": gp.config.String()})

	gp.buildDir = gp.args.HomeDir.Child(gp.config.Name)
	logger.Debug(Fields{"buildDir": gp.buildDir})

	if gp.args.ClearWorkspace {
		logger.Debug(Fields{"action": "delete", "path": gp.buildDir})
		err := os.RemoveAll(gp.buildDir.String())
		if err != nil {
			return nil, err
		}
		os.Exit(0)
	}

	must(os.MkdirAll(gp.args.BaseDir.String(), os.ModePerm))
	must(os.MkdirAll(gp.args.HomeDir.String(), os.ModePerm))
	must(os.MkdirAll(gp.buildDir.String(), os.ModePerm))

	res, err := gp.loadResources()
	if err != nil {
		return nil, err
	}
	gp.resources = res
	logger.Debug(Fields{"resources": gp.resources})

	gp.env = make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		gp.setEnv(pair[0], pair[1])
	}

	return gp, nil
}

// setEnv set a key/value environment variable
func (g *GoUp) setEnv(key string, val string) {
	g.env[key] = val
	logger.Debug(Fields{"$export": "env", key: val})
}

// loadResources only updates once a day or if the ~/.goup/resources.xml is missing
func (g *GoUp) loadResources() (*Resources, error) {
	file := g.args.HomeDir.Child("resources.xml")
	stat, err := os.Stat(file.String())
	if err != nil || time.Now().Sub(stat.ModTime()).Hours() > 24 {
		logger.Debug(Fields{"action": "downloading", "url": g.args.ResourcesUrl})
		_ = os.Remove(file.String())
		res, err := http.Get(g.args.ResourcesUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to get resource list: %v", err)
		}
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to download resource list: %v", err)
		}
		err = ioutil.WriteFile(file.String(), data, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to save resource list: %v", err)
		}
		logger.Debug(Fields{"action": "updated", "file": file})
	}
	res := &Resources{}
	logger.Debug(Fields{"action": "parsing", "file": file})
	err = res.Load(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load resources: %v", err)
	}
	return res, nil
}

// prepareAndroidSDK is required because the SDK is still not yet functional after downloading.
// Also it saves it things always in the wrong top level folder. Another wtf is that the sdkmanager
// only works with Java 8, even though we have Java 11 today.
func (g *GoUp) prepareAndroidSDK() error {
	sdkHome := Path(g.env["ANDROID_HOME"])
	if sdkHome.Child("platforms").Exists() {
		return nil
	}
	g.chdir(sdkHome.Child("bin"))
	_, err := g.Run2("./sdkmanager", []byte("y\n"), "platforms;android-28", "build-tools;28.0.3")
	if err != nil {
		return err
	}

	// just wtf: this version always writes the downloads into the wrong folder. I cannot get why, it does
	// so also in the systems commandline, so its broken at all?
	// This behavior does not make sense and every other tools expects it inside the sdk
	_ = os.Rename(g.toolchainPath().Child(".knownPackages").String(), sdkHome.Child(".knownPackages").String())
	_ = os.Rename(g.toolchainPath().Child("licenses").String(), sdkHome.Child("licenses").String())
	_ = os.Rename(g.toolchainPath().Child("platforms").String(), sdkHome.Child("platforms").String())
	_ = os.Rename(g.toolchainPath().Child("build-tools").String(), sdkHome.Child("build-tools").String())

	return nil
}

// prepareGomobileToolchain downloads go, ndk and sdk
func (g *GoUp) prepareGomobileToolchain() error {
	resources := make([]Resource, 0)

	// go
	goVersion := g.config.Build.Gomobile.Toolchain.Go
	if IsEmpty(goVersion) {
		goVersion = "1.12.4"
	}
	res, err := g.resources.Get("go", goVersion)
	if err != nil {
		return fmt.Errorf("cannot prepare android build: %v", err)
	}
	resources = append(resources, res)

	// android ndk
	ndkVersion := g.config.Build.Gomobile.Toolchain.Ndk
	if IsEmpty(ndkVersion) {
		ndkVersion = "r19c"
	}
	res, err = g.resources.Get("ndk", ndkVersion)
	if err != nil {
		return fmt.Errorf("cannot prepare android build: %v", err)
	}
	if g.hasAndroidBuild() {
		resources = append(resources, res)
	}

	// android sdk
	sdkVersion := g.config.Build.Gomobile.Toolchain.Sdk
	if IsEmpty(sdkVersion) {
		sdkVersion = "433796"
	}
	res, err = g.resources.Get("sdk", sdkVersion)
	if err != nil {
		return fmt.Errorf("cannot prepare android sdk: %v", err)
	}
	if g.hasAndroidBuild() {
		resources = append(resources, res)
	}

	// java jdk
	jdkVersion := g.config.Build.Gomobile.Toolchain.Jdk
	if IsEmpty(jdkVersion) {
		sdkVersion = "8u212b03"
	}
	res, err = g.resources.Get("jdk", jdkVersion)
	if err != nil {
		return fmt.Errorf("cannot prepare jdk: %v", err)
	}
	if g.hasAndroidBuild() {
		resources = append(resources, res)
	}

	for _, res := range resources {
		targetFolder := g.args.HomeDir.Child("toolchains").Child(res.Name + "-" + res.Version)
		if targetFolder.Exists() {
			logger.Debug(Fields{"toolchain": res.String(), "status": "exists"})
			continue
		}

		tmpTargetFolder := Path(targetFolder.String() + ".tmp")
		_ = os.RemoveAll(tmpTargetFolder.String())
		must(os.MkdirAll(tmpTargetFolder.String(), os.ModePerm))

		err := downloadAndUnpack(res.Url, tmpTargetFolder)
		if err != nil {
			return fmt.Errorf("failed to provide resource: %s: %v", res.String(), err)
		}

		// remove garbage
		for _, file := range tmpTargetFolder.List() {
			// remove hidden java virus files?
			if strings.HasPrefix(file.Name(), ".") && strings.Contains(file.Name(), "jdk") && !file.IsDir() {
				_ = os.Remove(file.String())
			}

			// finder trash stuff
			if file.Name() == ".DS_Store" {
				_ = os.Remove(file.String())
			}
		}

		files, err := ioutil.ReadDir(tmpTargetFolder.String())
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return fmt.Errorf("no files in resource: %s", res.String())
		}

		// just unwrap additional folder
		if len(files) == 1 && files[0].IsDir() {
			err := os.Rename(tmpTargetFolder.Child(files[0].Name()).String(), targetFolder.String())
			if err != nil {
				return err
			}
		} else {
			// already at root
			err := os.Rename(tmpTargetFolder.String(), targetFolder.String())
			if err != nil {
				return err
			}
		}

		_ = os.RemoveAll(tmpTargetFolder.String())

	}

	goRoot := g.args.HomeDir.Child("toolchains").Child("go-" + goVersion)
	javaHome := g.args.HomeDir.Child("toolchains").Child("jdk-" + jdkVersion).Child("Contents").Child("Home")
	sdkHome := g.args.HomeDir.Child("toolchains").Child("sdk-" + sdkVersion)

	g.setEnv("GOROOT", goRoot.String())
	g.setEnv("GOPATH", g.goPath().String())
	g.setEnv("PATH",
		goRoot.Child("bin").String()+":"+
			g.goPath().Child("bin").String()+":"+
			javaHome.Child("bin").String()+":"+
			sdkHome.String()+":"+
			sdkHome.Child("bin").String()+":"+
			g.env["PATH"])

	err = os.MkdirAll(g.goPath().String(), os.ModePerm)
	if err != nil {
		return err
	}

	_, _ = g.Run("which", "go")

	g.setEnv("ANDROID_NDK_HOME", g.args.HomeDir.Child("toolchains").Child("ndk-" + ndkVersion).String())
	g.setEnv("NDK_PATH", g.env["ANDROID_NDK_HOME"])
	g.setEnv("ANDROID_HOME", g.args.HomeDir.Child("toolchains").Child("sdk-" + sdkVersion).String())
	g.setEnv("ANDROID_SDK_ROOT", g.env["ANDROID_HOME"])

	g.setEnv("JAVA_HOME", javaHome.String())

	_, _ = g.Run("java", "-version")
	return nil
}

// goPath returns the artificial goPath
func (g *GoUp) goPath() Path {
	return g.buildDir.Child("go")
}

// toolchainPath returns the path for the toolchains
func (g *GoUp) toolchainPath() Path {
	return g.args.HomeDir.Child("toolchains")
}

// chdir changes the working directory of GoUp, especially it determines in which context external programs are
// executed
func (g *GoUp) chdir(path Path) {
	g.cwd = path
	logger.Debug(Fields{"cd": path})
}

// chmodX invokes chmod +x
func (g *GoUp) chmodX(path Path) error {
	_, err := g.Run("chmod", "+x", path.String())
	return err
}

func (g *GoUp) Run(name string, args ...string) ([]string, error) {
	return g.Run2(name, nil, args...)
}

func (g *GoUp) Run2(name string, pipeTo []byte, args ...string) ([]string, error) {
	cmd := exec.Command(name, args...)

	fields := Fields{}
	for k, v := range g.env {
		cmd.Env = append(cmd.Env, k+"="+v)
		fields[k] = v
	}
	logger.Debug(fields)

	tmpCmd := name + " "
	for _, a := range args {
		tmpCmd += a + " "
	}
	logger.Debug(Fields{"exec": tmpCmd})

	cmd.Dir = g.cwd.String()
	if len(pipeTo) != 0 {
		pipe, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		_, _ = pipe.Write(pipeTo)
	}

	stdoutStderr, err := cmd.CombinedOutput()

	lines := strings.Split(string(stdoutStderr), "\n")
	for _, line := range lines {
		if err != nil {
			logger.Error(Fields{"": line})
		} else {
			logger.Debug(Fields{"": line})
		}
	}

	return lines, err
}

// prepareGomobile installs gomobile into the gopath, if required
func (g *GoUp) prepareGomobile() error {
	if g.goPath().Child("bin").Child("gomobile").Exists() {
		return nil
	}
	g.chdir(g.goPath())
	g.setEnv("GO111MODULE", "off")

	logger.Debug(Fields{"action": "installing gomobile"})
	_, err := g.Run("go", "get", "-u", "golang.org/x/mobile/cmd/gomobile")
	if err != nil {
		return fmt.Errorf("failed to install gomobile: %v", err)
	}

	_, err = g.Run("bin/gomobile", "version")
	if err != nil {
		return fmt.Errorf("failed to invoke gomobile: %v", err)
	}

	// gomobile picks up ndk not anymore from -ndk but from ANDROID_NDK_HOME
	// also init actually does nothing anymore with prebuild toolchains, see also
	// https://github.com/golang/mobile/commit/ca80213619811c2fbed3ff8345accbd4ba924d45
	_, err = g.Run("bin/gomobile", "init")
	if err != nil {
		return fmt.Errorf("failed to init gomobile: %v", err)
	}
	return nil
}

// copyModulesToWorkspace performs the heavy lifting to get gomobile happy with "modules".
// It evaluates all module dependencies, collects them and copies the maximum resolved (by go mod vendor)
// version into the workspace
func (g *GoUp) copyModulesToWorkspace() error {
	dependencies := make(map[string]VendoredModule)
	g.chdir(g.goPath())
	g.setEnv("GO111MODULE", "on")
	resolvedLocalModulePaths := make([]Path, 0)
	for _, modPath := range g.config.Build.Gomobile.Modules {
		resolvedPath := Path(modPath).Resolve(g.args.BaseDir)

		//non-existing paths are treated as remote sources, they are downloaded directly
		if !resolvedPath.Exists() {
			// not a local mode, try to go get
			_, err := g.Run("go", "get", string(modPath))
			if err != nil {
				return err
			}
			resolvedPath = g.goPath().Child("pkg").Child("mod").Add(Path(modPath))
		}
		logger.Debug(Fields{"action": "processing", "path": resolvedPath})
		resolvedLocalModulePaths = append(resolvedLocalModulePaths, resolvedPath)

		modName, err := getModuleName(resolvedPath.Child("go.mod"))
		if err != nil {
			return fmt.Errorf("expected '%s' to have a go.mod file. This is not a go module: %v", resolvedPath, err)
		}
		logger.Debug(Fields{"name": modName})

		// copy declared go modules into go path
		targetDir := g.goPath().Child("src").Add(Path(modName))
		logger.Debug(Fields{"action": "removing", "path": targetDir})
		err = os.RemoveAll(targetDir.String())
		if err != nil {
			return fmt.Errorf("failed to clear directory %s: %v", targetDir, err)
		}
		logger.Debug(Fields{"action": "copying", "path": targetDir})

		err = CopyDir(resolvedPath.String(), targetDir.String())
		if err != nil {
			return fmt.Errorf("failed to copy directory %s: %v", targetDir, err)
		}

		// vendor module dependencies for each module
		g.chdir(targetDir)
		_, err = g.Run("go", "mod", "vendor")
		if err != nil {
			return fmt.Errorf("failed to vendor module dependencies: %v", err)
		}

		modules, err := ParseModulesTxT(targetDir.Child("vendor").Child("modules.txt").String())
		if err != nil {
			return fmt.Errorf("failed to parse vendor module information: %v", err)
		}

		// collected and inspect all modules: upgrade to the largest declared version, causing potential semver conflict
		for _, mod := range modules {
			logger.Debug(Fields{"action": "found", "module": mod.ModuleName, "version": mod.Version.String()})
			dep, ok := dependencies[mod.ModuleName]

			if !ok || mod.Version.IsNewer(dep.Version) {
				dependencies[mod.ModuleName] = mod
				dep = mod

				logger.Debug(Fields{"action": "upgrade", "module": mod.ModuleName, "version": mod.Version.String()})
			}
		}
	}

	// we collected all dependencies, now copy it into the workspace/gopath
	for _, dep := range dependencies {
		targetDir := g.goPath().Child("src").Add(Path(dep.ModuleName))
		err := os.RemoveAll(targetDir.Parent().String())
		if err != nil {
			return fmt.Errorf("failed to remove module target directory: %v", err)
		}
		err = os.MkdirAll(targetDir.Parent().String(), os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create module target directory: %v", err)
		}
		logger.Debug(Fields{"action": "move", "from": dep.Local, "to": targetDir})
		err = os.Rename(dep.Local.String(), targetDir.String())
		if err != nil {
			return fmt.Errorf("failed to move: %s->%s: %v", dep.Local, targetDir, err)
		}
	}

	// clear all vendor directories in copied modules
	for _, resolvedPath := range resolvedLocalModulePaths {
		modName, err := getModuleName(resolvedPath.Child("go.mod"))
		if err != nil {
			panic(err) // handled above already
		}
		targetDir := g.goPath().Child("src").Add(Path(modName)).Add("vendor")

		logger.Debug(Fields{"action": "remove", "file": targetDir})
		err = os.RemoveAll(targetDir.String())
		if err != nil {
			return fmt.Errorf("failed to remove: %s: %v", targetDir, err)
		}

	}
	return nil
}

// hasTargets checks if the target is defined
func (g *GoUp) hasTarget(target string) bool {
	for _, s := range g.args.Targets {
		if s == target || s == "all" {
			return true
		}
	}
	return false
}

// hasAndroidBuild returns true if a gomobile android section is defined and enabled
func (g *GoUp) hasAndroidBuild() bool {
	return g.config.Build.Gomobile != nil && g.config.Build.Gomobile.Android != nil && g.hasTarget("gomobile/android")
}

// hasIosBuild returns true if a gomobile ios section is defined and enabled
func (g *GoUp) hasIosBuild() bool {
	return g.config.Build.Gomobile != nil && g.config.Build.Gomobile.Ios != nil && g.hasTarget("gomobile/ios")
}

func (g *GoUp) compileGomobile() error {
	logger.Debug(Fields{"action": "compiling gomobile"})
	g.chdir(g.goPath())
	g.setEnv("GO111MODULE", "off")

	if g.hasAndroidBuild() {
		args := []string{"bind", "-v"}

		outFile := g.config.Build.Gomobile.Android.Out.Resolve(g.args.BaseDir)
		args = append(args, "-o", outFile.String())

		if len(g.config.Build.Gomobile.Android.Javapkg) > 0 {
			args = append(args, "-javapkg", g.config.Build.Gomobile.Android.Javapkg)
		}
		args = append(args, "-target=android")

		args = append(args, g.config.Build.Gomobile.Export...)
		_, err := g.Run("bin/gomobile", args...)
		if err != nil {
			return err
		}

	}

	if g.hasIosBuild() {
		args := []string{"bind", "-v"}

		if len(g.config.Build.Gomobile.Ios.Out) == 0 {
			g.config.Build.Gomobile.Ios.Out = Path("./" + g.config.Name + ".framework")
		}
		outFile := g.config.Build.Gomobile.Ios.Out.Resolve(g.args.BaseDir)
		args = append(args, "-o", outFile.String())

		if len(g.config.Build.Gomobile.Ios.Prefix) > 0 {
			args = append(args, "-prefix", g.config.Build.Gomobile.Ios.Prefix)
		}
		args = append(args, "-target=ios")

		args = append(args, g.config.Build.Gomobile.Export...)
		_, err := g.Run("bin/gomobile", args...)
		if err != nil {
			return err
		}
	}
	return nil
}

// isBuildRequired tries to detect if we need to build again. Because gomobile/cgo compiles really slow we want to
// avoid that in any case (e.g. 30s for "hello world" on a beefy machine) which takes a fraction of a second
// for go itself.
func (g *GoUp) isBuildRequired() bool {
	cacheFile := g.buildDir.Child("artifacts.json")
	g.artifactCache = &ArtifactCache{}
	err := g.artifactCache.Load(cacheFile.String())
	if err != nil {
		logger.Debug(Fields{"msg": "failed to load the build cache file, could be normal", "err": err.Error()})
		return true
	}

	inHash := g.calculateInHash()
	outHash := g.calculateOutHash()

	if g.artifactCache.InHash != inHash || g.artifactCache.OutHash != outHash {
		logger.Debug(Fields{"msg": "build cache indicates file changes"})
		return true
	}
	logger.Debug(Fields{"msg": "no need to build again"})
	return false
}

// updateBuildCache calculates and writes the current in/out hashes
func (g *GoUp) updateBuildCache() {
	inHash := g.calculateInHash()
	outHash := g.calculateOutHash()

	g.artifactCache.InHash = inHash
	g.artifactCache.OutHash = outHash
	cacheFile := g.buildDir.Child("artifacts.json")
	err := g.artifactCache.Save(cacheFile.String())
	if err != nil {
		fmt.Println(err)
	}
}

// Build performs the actual build process
func (g *GoUp) Build() error {
	started := time.Now()
	defer func() {
		logger.Info(Fields{"msg": "build done", "time": time.Now().Sub(started).String()})
	}()
	if !g.isBuildRequired() {
		return nil
	}

	err := g.prepareGomobileToolchain()
	if err != nil {
		return fmt.Errorf("failed to prepare gomobile build: %v", err)
	}

	err = g.prepareGomobile()
	if err != nil {
		return err
	}

	err = g.prepareAndroidSDK()
	if err != nil {
		return fmt.Errorf("failed to init android sdk: %v", err)
	}

	err = g.copyModulesToWorkspace()
	if err != nil {
		return err
	}

	err = g.compileGomobile()
	if err != nil {
		return err
	}

	g.updateBuildCache()
	return nil
}
