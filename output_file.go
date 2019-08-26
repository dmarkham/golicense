package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/mitchellh/golicense/config"
	"github.com/mitchellh/golicense/license"
	"github.com/mitchellh/golicense/module"
)

// FileOutput is an Output implementation that outputs to the terminal.
type FileOutput struct {

	// Config is the configuration (if any). This will be used to check
	// if a license is allowed or not.
	Config *config.Config

	// Modules is the full list of modules that will be checked. This is
	// optional. If this is given in advance, then the output will be cleanly
	// aligned.
	Modules []module.Module

	modules   map[string]string
	moduleMax int
	exitCode  int
	lineMax   int
	once      sync.Once
	lock      sync.Mutex
}

func (o *FileOutput) ExitCode() int {
	return o.exitCode
}

// Start implements Output
func (o *FileOutput) Start(m *module.Module) {
	o.once.Do(o.init)
	return
}

// Update implements Output
func (o *FileOutput) Update(m *module.Module, t license.StatusType, msg string) {
	o.once.Do(o.init)
	return
}

// Finish implements Output
func (o *FileOutput) Finish(m *module.Module, l *license.License, err error) {
	o.once.Do(o.init)

	if o.Config == nil {
		fmt.Printf("We need a config to run")
		os.Exit(1)
	}

	state := o.Config.Allowed(l)
	switch state {
	case config.StateAllowed:
	case config.StateDenied, config.StateUnknown:
		fmt.Printf("Denied or Unknown Licence: %v -> (%v)\n", m.String(), l.String())
		os.Exit(1)
		//o.exitCode = 1
	}

	// on a Override we have to go find it or it's will error out
	if o.Config.Override[m.Path] != "" {
		oPath := o.Config.OverridePath + "/" + m.Path + "/" + "LICENSE"
		f, err := os.Open(oPath)
		if err == nil {
			b, err := ioutil.ReadAll(f)
			if err == nil {
				l.Text = string(b)
			} else {
				fmt.Println("Error Reading from Path: " + oPath)
			}
		} else {
			fmt.Println("Nothing found for " + oPath)
		}
	}

	//fmt.Printf(
	//	fmt.Sprintf("%s\t%s\t%s\n",
	//		o.paddedModule(m),
	//		l.String(),
	//		base64.StdEncoding.EncodeToString([]byte(l.Text))))

	if len(l.Text) > 0 {
		dir, _ := path.Split(o.Config.ExecPath)
		dir = dir + "License/" + m.Path + "/"

		err := os.MkdirAll(dir, 0777)
		if err != nil {
			panic(err)
		}
		fh, err := os.Create(dir + "LICENSE")
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(fh, strings.NewReader(l.Text))
		if err != nil {
			panic(err)
		}
		err = fh.Close()
		if err != nil {
			panic(err)
		}

	} else {
		fmt.Printf("Missing Licence: %v -> (%v)\n", m.String(), l.String())
		os.Exit(1)
	}

	return
}

// Close implements Output
func (o *FileOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	return nil
}

// paddedModule returns the name of the module padded so that they align nicely.
func (o *FileOutput) paddedModule(m *module.Module) string {
	o.once.Do(o.init)

	if o.moduleMax == 0 {
		return m.Path
	}

	// Pad the path so that it is equivalent to the moduleMax length
	return m.Path + strings.Repeat(" ", o.moduleMax-len(m.Path))
}

func (o *FileOutput) init() {
	if o.modules == nil {
		o.modules = make(map[string]string)
	}

	// Calculate the maximum module length
	for _, m := range o.Modules {
		if v := len(m.Path); v > o.moduleMax {
			o.moduleMax = v
		}
	}

}
