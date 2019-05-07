// Copyright 2019 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// A ModuleSpecifier is either a relative path (e.g. ./my/module) or absolute (e.g. /home/usr/my/mod)
// or remote (e.g. github.com/my/mod).
type ModuleSpecifier string

// A GoModuleName is either an actual name of a go module as defined in go.mod
// or a package name in a GOPATH
type GoModuleName string

// A GoUpConfiguration is parsed from the goup.yaml file and specifies various build configurations.
type GoUpConfiguration struct {
	// The Name is used to setup a custom workspace and tools.
	// You should not invoke parallel builds for the same project
	Name string

	// The build section defines what and how goup should work
	Build *Build
}

// The Build section defines what and how goup should work
type Build struct {
	Gomobile *BuildGomobile
}

// The BuildGomobile build, e.g. for ios or android
type BuildGomobile struct {
	// the toolchain section is required to setup a stable gomobile building experience
	Toolchain BuildGomobileToolchain
	// The ios section defines how our iOS library is build. This only works on MacOS with XCode installed
	Ios *Ios
	// The android section defines how our android build is executed
	Android *Android

	Modules []ModuleSpecifier

	// The export section defines all exported packages which are passed to gobind by gomobile.
	// Gomobile does not generate transitives exports, so you need to declare all
	// packages containing types and methods which you want to have bindings for.
	// Be careful with name conflicts, because the last part of the package will be used
	// to scope the types.
	Export []string
}

// The BuildGomobileToolchain section is required to setup a stable gomobile building experience
type BuildGomobileToolchain struct {
	// which go version? e.g. 1.12.4
	Go string
	// which android ndk version? e.g. r19c
	Ndk string
	// which android sdk version? e.g. 4333796
	Sdk string
	// which java jdk version? e.g. 8u212b03
	Jdk string
}

// The Ios section defines how our iOS library is build. This only works on MacOS with XCode installed
type Ios struct {
	// The gomobile -prefix flag
	Prefix string
	// The gomobile -o flag, this will be a folder
	Out Path
	// The gomobile -bundleid flag sets the bundle ID to use with the app.
	Bundleid string
	// The gomobile -ldflags flag
	Ldflags string
	// The disabled flag can be used to declare but disable this build
	Disabled bool
}

// The Android section defines how our android build is executed
type Android struct {
	// The gomobile -javapkg flag prefixes the generated packages
	Javapkg string
	// The gomobile -o flag, this will be an aar file
	Out Path
	// The gomobile -ldflags flag
	Ldflags string
}

// Load reads a build.yaml file into the receiver
func (c *GoUpConfiguration) Load(file Path) error {
	data, err := ioutil.ReadFile(file.String())
	if err != nil {
		return fmt.Errorf("failed to load GoUpConfiguration from %s: %v", file, err)
	}
	err = yaml.Unmarshal([]byte(data), c)
	if err != nil {
		return fmt.Errorf("failed to parse GoUpConfiguration from %s: %v", file, err)
	}
	return nil
}

func (c *GoUpConfiguration) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}
