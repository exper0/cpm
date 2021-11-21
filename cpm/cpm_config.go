package cpm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

const (
	OnExitTerminate					= "terminate"
	OnExitTerminateIfError			= "terminate_if_error"
	OnExitDoNothing					= "do_nothing"
	DefaultOnExit 					= OnExitTerminate

	TerminationModePropagate    	= "propagate"
	TerminationModeIgnoreErrors 	= "ignore_errors"
	DefaultTerminationMode      	= TerminationModePropagate

	StopViaCommand    				= "command"
	StopViaSignal     				= "signal"
	DefaultStopMethod 				= StopViaCommand
	DefaultStopCommand 				= "kill {{ .Pid}}"
	DefaultStopTimeout				= 2
	DefaultStopSignal				= "SIGTERM"
)

type ProcessAttributes struct {
	StartCommand 		string `json:"start_command"`
	WorkDirectory		string `json:"work_directory"`
	PidFile 			string `json:"pid_file"`
	OnExit				string `json:"on_exit"`
	OnExitCommand		string `json:"on_exit_command"`
	TerminationMode 	string `json:"termination_mode"`
	StopVia         	string `json:"stop_via"`
	StopCommand     	string `json:"stop_command"`
	StopSignal 			string `json:"stop_signal"`
	StopTimeout			uint8  `json:"stop_timeout"`
}

type Process struct {
	Name		string 				`json:"name"`
	Pid 		uint16      		// not from json config
	ExitCode	uint8				// not from json config
	Attributes 	ProcessAttributes 	`json:"attributes"`
}

type Config struct {
	GlobalAttributes 	ProcessAttributes 	`json:"global_attributes"`
	Processes  [] 		Process     		`json:"processes"`
}

func (config *Config) Read(fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &config); err != nil {
		return err
	}
	config.merge()
	if err = config.validate(); err != nil {
		return err
	}
	return nil
}

func (config *Config) merge() {
	if config.GlobalAttributes.OnExit == "" {
		config.GlobalAttributes.OnExit = DefaultOnExit
	}
	if config.GlobalAttributes.TerminationMode == "" {
		config.GlobalAttributes.TerminationMode = DefaultTerminationMode
	}
	if config.GlobalAttributes.StopVia == "" {
		config.GlobalAttributes.StopVia = DefaultStopMethod
	}
	if config.GlobalAttributes.StopCommand == "" {
		config.GlobalAttributes.StopCommand = DefaultStopCommand
	}
	if config.GlobalAttributes.StopSignal == "" {
		config.GlobalAttributes.StopSignal = DefaultStopSignal
	}
	if config.GlobalAttributes.StopTimeout == 0 {
		config.GlobalAttributes.StopTimeout = DefaultStopTimeout
	}
	for _, process := range config.Processes {
		if process.Attributes.OnExit == "" {
			process.Attributes.OnExit = config.GlobalAttributes.OnExit
		}
		if process.Attributes.TerminationMode == "" {
			process.Attributes.OnExitCommand = config.GlobalAttributes.TerminationMode
		}
		if process.Attributes.StopVia == "" {
			process.Attributes.StopVia = config.GlobalAttributes.StopVia
		}
		if process.Attributes.StopCommand == "" {
			process.Attributes.StopCommand = config.GlobalAttributes.StopCommand
		}
		if process.Attributes.StopSignal == "" {
			process.Attributes.StopSignal = config.GlobalAttributes.StopSignal
		}
		if process.Attributes.StopTimeout == 0 {
			process.Attributes.StopTimeout = config.GlobalAttributes.StopTimeout
		}
	}
}

func (config *Config) validate() error {
	for _, process := range config.Processes {
		if process.Name == "" {
			return errors.New("name cannot be empty")
		}
		if process.Attributes.StartCommand == "" {
			return fmt.Errorf("process '%s': start command cannot be empty", process.Name)
		}
		if process.Attributes.StopVia == "" {
			return fmt.Errorf("process '%s': stop method cannot be empty", process.Name)
		}
		if process.Attributes.StopVia == StopViaSignal && process.Attributes.StopSignal == "" {
			return fmt.Errorf("process '%s': stop via signal is chosen, StopSignal cannot be empty", process.Name)
		}
		if process.Attributes.StopVia == StopViaCommand && process.Attributes.StopCommand == "" {
			return fmt.Errorf("process '%s': stop via command is chosen, StopComsmand cannot be empty", process.Name)
		}
		if process.Attributes.OnExit == "" {
			return fmt.Errorf("process '%s': OnExit cannot be empty", process.Name)
		}
		if process.Attributes.TerminationMode == "" {
			return fmt.Errorf("process '%s': TerminationMode cannot be empty", process.Name)
		}
	}
	return nil
}
