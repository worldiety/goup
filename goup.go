package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Goup contains the actual state of the goup program
type Goup struct {
	// The program arguments
	args *Args
	// The parsed config
	config *GoUpConfiguration

	// the buildDir is the folder where we collect everything for this project
	buildDir Path

	resources *Resources
}

// NewGoupBuilder creates a new Goup builder
func NewGoup(args *Args) (*Goup, error) {
	gp := &Goup{}
	gp.args = args
	gp.config = &GoUpConfiguration{}
	err := gp.config.Load(gp.args.BuildFile)
	if err != nil {
		return nil, err
	}

	logger.Debug(Fields{"buildFile": gp.config.String()})

	gp.buildDir = gp.args.HomeDir.Child(gp.config.Name)
	logger.Debug(Fields{"buildDir": gp.buildDir})

	must(os.MkdirAll(gp.args.BaseDir.String(), os.ModePerm))
	must(os.MkdirAll(gp.args.HomeDir.String(), os.ModePerm))
	must(os.MkdirAll(gp.buildDir.String(), os.ModePerm))

	res, err := gp.loadResources()
	if err != nil {
		return nil, err
	}
	gp.resources = res
	logger.Debug(Fields{"resources": gp.resources})
	return gp, nil
}

// loadResources only updates once a day or if the ~/.goup/resources.xml is missing
func (g *Goup) loadResources() (*Resources, error) {
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

// prepareGomobileToolchain downloads go, ndk and sdk
func (g *Goup) prepareGomobileToolchain() error {
	if !g.config.HasAndroidBuild() {
		return nil
	}
	resources := make([]Resource, 0)

	goVersion := g.config.Build.Gomobile.Toolchain.Go
	if IsEmpty(goVersion) {
		goVersion = "1.12.4"
	}
	res, err := g.resources.Get("go", goVersion)
	if err != nil {
		return fmt.Errorf("cannot prepare android build: %v", err)
	}
	resources = append(resources, res)

	ndkVersion := g.config.Build.Gomobile.Toolchain.Ndk
	if IsEmpty(ndkVersion) {
		ndkVersion = "r19c"
	}
	res, err = g.resources.Get("ndk", ndkVersion)
	if err != nil {
		return fmt.Errorf("cannot prepare android build: %v", err)
	}
	resources = append(resources, res)

	sdkVersion := g.config.Build.Gomobile.Toolchain.Sdk
	if IsEmpty(sdkVersion) {
		sdkVersion = "433796"
	}
	res, err = g.resources.Get("sdk", sdkVersion)
	if err != nil {
		return fmt.Errorf("cannot prepare android build: %v", err)
	}
	resources = append(resources, res)

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
	}

	return nil
}

// Build performs the actual build process
func (g *Goup) Build() error {
	err := g.prepareGomobileToolchain()
	if err != nil {
		return fmt.Errorf("failed to prepare gomobile build: %v", err)
	}
	return nil
}
