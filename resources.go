package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"runtime"
)

// A Resource has a unique combination of name, version, os and architecture.
type Resource struct {
	// Name of the resource e.g. go
	Name string

	// Url to download, e.g. https://dl.google.com/go/go1.12.4.darwin-amd64.tar.gz
	Url string

	// Version of the resource e.g. 1.12.4
	Version string

	// OS e.g. darwin
	OS string

	// Arch e.g. amd64
	Arch string
}

// Resources is just a slice of resources
type Resources []Resource

// Get loops over all resources and returns the fitting resource for the current os/arch combination
func (r *Resources) Get(name string, version string) (Resource, error) {
	for _, e := range r {
		if e.Name == name && e.Version == version && e.Arch == runtime.GOARCH && e.OS == runtime.GOOS {
			return e, nil
		}
	}
	return Resource{}, fmt.Errorf("no such resource: %s@%s for os=%s and arch=%s", name, version, runtime.GOOS, runtime.GOARCH)
}

// Load parses a local xml file and replaces the contents of resources
func (r *Resources) Load(fname Path) error {
	*r = make([]Resource, 0)
	data, err := ioutil.ReadFile(fname.String())
	if err != nil {
		return fmt.Errorf("unable to read xml file: %v", err)
	}
	err = xml.Unmarshal(data, r)
	if err != nil {
		return fmt.Errorf("unable to parse xml: %v", err)
	}

	return nil
}
