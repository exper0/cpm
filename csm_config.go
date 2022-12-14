package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type ServiceConfig struct {
	//  Mandatory parameters
	Name         string `json:"name"`
	StartCommand string `json:"start_command"`
	// optional parameters
	WorkDirectory string `json:"work_directory"`
	StartupDelay  uint   `json:"startup_delay"`
	OnExit        string `json:"on_exit"`
	MaxRestarts   uint   `json:"max_restarts"`
	StopSignalStr string `json:"stop_signal"`
	Stdout        string `json:"stdout"`
	Stderr        string `json:"stderr"`

	// parsed parameters
	CommandLine []string
	StopSignal  syscall.Signal

	// used for tests
	ExpectedStdOut      string `json:"expected_std_out"`
	ExpectedStdErr      string `json:"expected_std_err"`
	ExpectedReturnValue int    `json:"expected_return_value"`
}

type Config struct {
	ShutdownTimeoutSec  uint8           `json:"shutdown_timeout_sec"`
	Services            []ServiceConfig `json:"services"`
	ExpectedReturnValue int             `json:"expected_return_value"`
}

func NewConfig(fileName string) (*Config, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var config Config
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	config.setDefaults()
	if err = config.parse(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (config *Config) setDefaults() {
	if config.ShutdownTimeoutSec == 0 {
		config.ShutdownTimeoutSec = DefaultShutdownTimout
	}
	for i := 0; i < len(config.Services); i++ {
		s := &config.Services[i]
		if s.OnExit == "" {
			s.OnExit = DefaultOnExit
		}
		if s.StopSignalStr == "" {
			s.StopSignalStr = DefaultStopSignal
		}
		if s.Stdout == "" {
			s.Stdout = Stdout
		}
		if s.Stderr == "" {
			s.Stderr = Stderr
		}
		if s.StartupDelay == 0 {
			s.StartupDelay = DefaultStartupDelay
		}
	}
}

func (config *Config) parse() error {
	if len(config.Services) == 0 {
		return errors.New("there should be at least one service")
	}
	for i := 0; i < len(config.Services); i++ {
		s := &config.Services[i]
		if s.Name == "" {
			return fmt.Errorf("service %d: name cannot be empty", i+1)
		}
		if startCmd := strings.Trim(s.StartCommand, " "); startCmd == "" {
			return fmt.Errorf("service '%s': start command cannot be empty", s.Name)
		} else {
			s.CommandLine = strings.Split(startCmd, " ")
		}
		if onExit := strings.Trim(s.OnExit, " "); onExit == "" {
			return fmt.Errorf("service '%s': OnExit cannot be empty", s.Name)
		} else {
			s.OnExit = onExit
		}
		if _, ok := onExitTable[s.OnExit]; !ok {
			return fmt.Errorf("service '%s': unknown value for on_exit parameter: '%s'", s.Name, s.OnExit)
		}
		stopSignal := strings.Trim(s.StopSignalStr, " ")
		if val, ok := signals[stopSignal]; ok {
			s.StopSignalStr = stopSignal
			s.StopSignal = val
		} else {
			return fmt.Errorf("service '%s': unknown value for stop_signal parameter: '%s'", s.Name, stopSignal)
		}
		if s.Stdout != Stdout || s.Stdout != s.Stderr {
			if v, err := validPath(s.Stdout); !v {
				return fmt.Errorf("invalid value(%s) for stdout property for %s: %w", s.Stdout, s.Name, err)
			}
		}
		if s.Stderr != Stdout || s.Stderr != s.Stderr {
			if v, err := validPath(s.Stderr); !v {
				return fmt.Errorf("invalid value(%s) for stdout property for %s: %w", s.Stderr, s.Name, err)
			}
		}
	}
	return nil
}

func validPath(p string) (bool, error) {
	if _, err := os.Stat(filepath.Dir(p)); err != nil {
		return false, err
	}
	return true, nil
}

const (
	OnExitTerminate              = "terminate"
	OnExitTerminateIfError       = "terminate_if_error"
	OnExitDoNothing              = "do_nothing"
	OnExitRestart                = "restart"
	DefaultOnExit                = OnExitTerminate
	DefaultStopSignal            = "SIGTERM"
	DefaultShutdownTimout  uint8 = 3
	DefaultStartupDelay    uint  = 1000
	Stdout                       = "stdout"
	Stderr                       = "stderr"
)

// Signal table
var signals = map[string]syscall.Signal{
	"SIGABRT":   syscall.SIGABRT,
	"SIGALRM":   syscall.SIGALRM,
	"SIGBUS":    syscall.SIGBUS,
	"SIGCHLD":   syscall.SIGCHLD,
	"SIGCONT":   syscall.SIGCONT,
	"SIGEMT":    syscall.SIGEMT,
	"SIGFPE":    syscall.SIGFPE,
	"SIGHUP":    syscall.SIGHUP,
	"SIGILL":    syscall.SIGKILL,
	"SIGINFO":   syscall.SIGINFO,
	"SIGINT":    syscall.SIGINT,
	"SIGIO":     syscall.SIGIO,
	"SIGIOT":    syscall.SIGIOT,
	"SIGKILL":   syscall.SIGKILL,
	"SIGPIPE":   syscall.SIGPIPE,
	"SIGPROF":   syscall.SIGPROF,
	"SIGQUIT":   syscall.SIGQUIT,
	"SIGSEGV":   syscall.SIGSEGV,
	"SIGSTOP":   syscall.SIGSTOP,
	"SIGSYS":    syscall.SIGSYS,
	"SIGTERM":   syscall.SIGTERM,
	"SIGTRAP":   syscall.SIGTRAP,
	"SIGTSTP":   syscall.SIGTSTP,
	"SIGTTIN":   syscall.SIGTTIN,
	"SIGTTOU":   syscall.SIGTTOU,
	"SIGURG":    syscall.SIGURG,
	"SIGUSR1":   syscall.SIGUSR1,
	"SIGUSR2":   syscall.SIGUSR2,
	"SIGVTALRM": syscall.SIGVTALRM,
	"SIGWINCH":  syscall.SIGWINCH,
	"SIGXCPU":   syscall.SIGXCPU,
	"SIGXFSZ":   syscall.SIGXFSZ,
}

var onExitTable = map[string]bool{
	OnExitDoNothing:        true,
	OnExitRestart:          true,
	OnExitTerminate:        true,
	OnExitTerminateIfError: true,
}
