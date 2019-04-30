package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Build is the model for the gomobile.build file
type Build struct {
	// Project is the (globally) unique name. It is used to locate a unique and recycled gopath
	Project string `json:"project"`
	// IOS contains the specific target build config
	IOS *IOS `json:"ios"`
	// Android contains the specific target build config
	Android *Android `json:"android"`
	// Imports refer to local absolute or relative paths
	Imports []Path `json:"import"`
	// Exports refers to gopath specific full qualified package names
	Exports []string `json:"export"`
}

// IOS covers the iOS part of the build section
type IOS struct {
	//Prefix defines the -prefix value for gomobile
	Prefix string `json:"prefix"`
	//Out defines the -o value for gomobile
	Out string `json:"prefix"`
	//Ldflags defines the -ldflags for gomobile
	Ldflags string `json:"ldflags"`
}

// Android covers the Android part of the build section
type Android struct {
	//Package defines the -javapkg value for gomobile
	Package string `json:"pkg"`
	//Out defines the -o value for gomobile
	Out string `json:"prefix"`
	//Ldflags defines the -ldflags for gomobile
	Ldflags string `json:"ldflags"`
}

func (b *Build) String() string {
	data, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// LoadBuildFile tries to load a gomobile.build file in json format from the given filename
func LoadBuildFile(filename Path) (*Build, error) {
	dst := &Build{}
	data, err := ioutil.ReadFile(filename.String())
	if err != nil {
		return nil, fmt.Errorf("unable to load build file: %v", err)
	}
	err = json.Unmarshal(data, dst)
	if err != nil {
		return nil, fmt.Errorf("unable to parse build file %s: %v", filename, err)
	}
	return dst, err
}
