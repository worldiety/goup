package main

import (
	"io/ioutil"
	"os"
)

// ReadVersion tries to read the entire file as a string
func ReadVersion(fname string) string {
	str, _ := ioutil.ReadFile(fname)
	return string(str)
}

// WriteVersion overwrites the denoted file with the version string
func WriteVersion(fname string, version string) {
	_ = ioutil.WriteFile(fname, []byte(version), os.ModePerm)
}
