package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), "case") {
			fmt.Println("found test case " + f.Name())
			err := os.Chdir(filepath.Join("testdata", f.Name()))
			if err != nil {
				t.Fatal(err)
			}
			t.Run(f.Name(), func(t *testing.T) {
				if config, err := NewConfig("config.json"); err == nil {
					ctr := NewServiceManager(config)
					if err := ctr.StartAll(); err != nil {
						t.Fatal(err)
					}
					smr := ctr.MainLoop()
					if smr != config.ExpectedReturnValue {
						t.Errorf("expected %d but got %d", config.ExpectedReturnValue, smr)
					}
					for i := 0; i < len(ctr.services); i++ {
						s := &ctr.services[i]
						expectedStdOut := (*s).config.ExpectedStdOut
						if expectedStdOut != "" {
							actualStdOutB, err := ioutil.ReadFile((*s).config.Stdout)
							actualStdOut := string(actualStdOutB)
							if err == nil {
								if strings.Trim(actualStdOut, "\n") != strings.Trim(expectedStdOut, "\n") {
									t.Errorf("stdout content is different: expected '%s' but got '%s'", expectedStdOut, actualStdOut)
								}
							} else {
								t.Fatal(err)
							}
						}
					}
				} else {
					t.Fatal(err)
				}
			})
		}
	}
}
