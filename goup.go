package main

import (
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

	return gp, nil
}

// updateResourceList only updates once a day or if the ~/.goup/resources.xml is missing
func (g *Goup) updateResourceList() error {
	file := g.args.HomeDir.Child("resources.xml")
	stat, err := os.Stat(file.String())
	if err != nil || time.Now().Sub(stat.ModTime()).Hours() > 24 {
		_ = os.Remove(file.String())
		http.Get()
	}
}
