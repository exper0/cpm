package main

import (
	"fmt"
	"os"
)

type Service struct {
	process  *os.Process
	exitCode *int
	stopped  bool
	restarts uint
	stdout   *os.File
	stderr   *os.File
	config   *ServiceConfig
}

func NewService(config *ServiceConfig) *Service {
	return &Service{config: config}
}

func (service *Service) Start() error {
	var wd string
	if service.config.WorkDirectory == "" {
		if _wd, err := os.Getwd(); err == nil {
			wd = _wd
		} else {
			return err
		}
	} else {
		wd = service.config.WorkDirectory
	}
	//var stdout *os.File
	//var stderr *os.File
	var err error
	service.stdout, err = service.getOutput(service.config.Stdout)
	if err != nil {
		return fmt.Errorf("unable to prepare stdout: %w", err)
	}
	service.stderr, err = service.getOutput(service.config.Stderr)
	if err != nil {
		service.closeOutputs()
		return fmt.Errorf("unable to prepare stderr: %w", err)
	}
	p, err := os.StartProcess(service.config.CommandLine[0], service.config.CommandLine, &os.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, service.stdout, service.stderr},
	})
	if err != nil {
		service.closeOutputs()
		return fmt.Errorf("unable to run '%s': %w", service.config.StartCommand, err)
	}
	service.process = p
	//if service.config.StartupDelay > 0 {
	//	time.Sleep(time.Duration(service.config.StartupDelay) * time.Millisecond)
	//}
	return nil
}

func (service *Service) getOutput(s string) (*os.File, error) {
	if s == Stdout {
		return os.Stdout, nil
	}
	if s == Stderr {
		return os.Stderr, nil
	}
	//var f *os.File
	fi, err := os.Stat(s)
	if fi != nil {
		if f, err := os.OpenFile(s, os.O_WRONLY|os.O_APPEND, 0644); err != nil {
			return nil, fmt.Errorf("unable to open output file for path %s: %w", s, err)
		} else {
			return f, nil
		}
	}
	if os.IsNotExist(err) {
		if f, err := os.Create(s); err != nil {
			return nil, fmt.Errorf("unable to create output file for path %s: %w", s, err)
		} else {
			return f, nil
		}
	}
	return nil, fmt.Errorf("unable to open output file for path %s: %w", s, err)
}

func (service *Service) closeOutputs() {
	if service.stdout != nil && service.stdout != os.Stdout {
		service.stdout.Close()
	}
	if service.stderr != nil && service.stderr != os.Stderr {
		service.stderr.Close()
	}
}
