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

// A BuildCacheProject contains information about a build, to decide if it should get compiled again.
// Actually these are just the program arguments.
//
// Calculating the hash is exact but very expensive, we should use a file tree with last mod time stamps.
type BuildCache struct {
	WorkingDir Path
	BuildFile  Path
	HomeDir    Path

	// The sha sum of all input files
	InHash string

	// The sha sum of the output
	OutHash string
}

// calculateInHash takes all input files, like *.go and go.mod and gomobile.build. We do not want to check
// remote changes, which may cause unstable builds, but that is a user problem, not ours.
func (b *Builder) calculateInHash() string {
	extensions := []string{".go", "go.mod", "go.sum", "gomobile.build"}
	hasher := sha1.New()
	hasher.Write(sloppyBytes(ioutil.ReadFile(b.BuildFile.String())))
	for _, modPath := range b.BuildConfig.Imports {
		resolvedPath := modPath.Resolve(b.BaseDir)
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
				if b.Verbose {
					b.PP("build cache: %s", file)
				}

				hasher.Write(sloppyBytes(ioutil.ReadFile(file)))
			}
		}

	}
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	if b.Verbose {
		b.PP("hash of input files is: %s", hash)
	}
	return hash
}

// calculateOutHash takes the generated out files to hash them. Absent files are hashed as 0 bytes.
func (b *Builder) calculateOutHash() string {
	hasher := sha1.New()
	if b.BuildConfig.Build.Android != nil {
		outFile := b.BuildConfig.Build.Android.Out.Resolve(b.BaseDir)
		hasher.Write(sloppyBytes(ioutil.ReadFile(outFile.String())))
	}
	if b.BuildConfig.Build.IOS != nil {
		//TODO is this a file or folder?
		outFile := b.BuildConfig.Build.IOS.Out.Resolve(b.BaseDir)
		hasher.Write(sloppyBytes(ioutil.ReadFile(outFile.String())))
	}
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	if b.Verbose {
		b.PP("hash of output files is: %s", hash)
	}
	return hash
}

// Save serializes the cache data into json
func (b *BuildCache) Save(fname string) error {
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

// Loads deserializes the cache data from json and ensures that the cache is cleared.
func (b *BuildCache) Load(fname string) error {
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
