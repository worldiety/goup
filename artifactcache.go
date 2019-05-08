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
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// A ArtifactCache contains information about a build, to decide if it should get compiled again.
// Actually these are just the program arguments.
//
// Calculating the hash is exact but very expensive, we should use a file tree with last mod time stamps.
type ArtifactCache struct {
	// The sha sum of all input files
	InHash string

	// The sha sum of the output
	OutHash string
}

// calculateInHash takes all input files, like *.go and go.mod and gomobile.build. We do not want to check
// remote changes, which may cause unstable builds, but that is a user problem, not ours.
func (g *GoUp) calculateInHash() string {
	extensions := []string{".go", "go.mod", "go.sum"}
	hasher := sha1.New()
	hasher.Write(sloppyBytes(ioutil.ReadFile(g.args.BuildFile.String())))
	for _, modPath := range g.config.Build.Gomobile.Modules {
		resolvedPath := Path(modPath).Resolve(g.args.BaseDir)
		if !resolvedPath.Exists() {
			// this is the external case, we will never check that for performance reasons
			continue
		}
		files, err := ListFiles(resolvedPath.String())
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			acceptable := false
			for _, ext := range extensions {
				if strings.HasSuffix(file, ext) {
					acceptable = true
					break
				}
			}
			if acceptable {
				logger.Debug(Fields{"in-artifact": file})

				hasher.Write(sloppyBytes(ioutil.ReadFile(file)))
			}
		}

	}
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	logger.Debug(Fields{"in-hash": hash})
	return hash
}

// calculateOutHash takes the generated out files to hash them. Absent files are hashed as 0 bytes.
func (g *GoUp) calculateOutHash() string {
	extensions := []string{".h", ".plist", ".modulemap"}
	hasher := sha1.New()
	if g.hasAndroidBuild() {
		outFile := g.config.Build.Gomobile.Android.Out.Resolve(g.args.BaseDir)
		// we need to hash the file, to avoid rebuilding failures when switching android build on and the dir is missing
		hasher.Write([]byte(outFile.String()))
		hasher.Write(sloppyBytes(ioutil.ReadFile(outFile.String())))
	}
	if g.hasIosBuild() {
		outFolder := g.config.Build.Gomobile.Ios.Out.Resolve(g.args.BaseDir)
		// we need to hash the folder, to avoid rebuilding failures when switching ios build on and the dir is missing
		hasher.Write([]byte(outFolder.String()))
		files, err := ListFiles(outFolder.String())
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			acceptable := false
			for _, ext := range extensions {
				if strings.HasSuffix(file, ext) {
					acceptable = true
					break
				}
			}
			if acceptable {
				logger.Debug(Fields{"in-artifact": file})

				hasher.Write(sloppyBytes(ioutil.ReadFile(file)))
			}
		}
	}
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	logger.Debug(Fields{"out-hash": hash})
	return hash
}

// Save serializes the cache data into json
func (b *ArtifactCache) Save(fname string) error {
	data, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("failed to marshal: %v", err)
	}
	err = ioutil.WriteFile(fname, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to save: %v", err)
	}
	return nil
}

// Load deserializes the cache data from json and ensures that the cache is cleared.
func (b *ArtifactCache) Load(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("failed to load json: %v", err)
	}
	err = json.Unmarshal(data, b)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}
	return nil
}

func sloppyBytes(data []byte, err error) []byte {
	if err != nil {
		fmt.Println(err)
	}
	return data
}
