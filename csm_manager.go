package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ServiceManager struct {
	config   *Config
	services []*Service
	started  bool
}

func NewServiceManager(config *Config) *ServiceManager {
	var sm ServiceManager
	sm.config = config
	sm.services = make([]*Service, len(config.Services))
	for i := 0; i < len(config.Services); i++ {
		sm.services[i] = NewService(&config.Services[i])
	}
	return &sm
}

func (sm *ServiceManager) StartAll() error {
	if sm.started {
		return errors.New("already started")
	}
	for _, s := range sm.services {
		if err := s.Start(); err != nil {
			return fmt.Errorf("unable to start service %s: '%w'. Exitting",
				s.config.Name, err)
		} else {
			log.Printf("service %s(%s) successfully started", s.config.Name, s.config.StartCommand)
		}
	}
	sm.started = true
	return nil
}

func (sm *ServiceManager) MainLoop() int {
	if !sm.started {
		log.Printf("services are not started")
		return -1
	}
	var quitChan = make(chan os.Signal, 1)
	var childChan = make(chan os.Signal, len(sm.services))
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(childChan, syscall.SIGCHLD)
loop:
	for {
		select {
		case sig := <-quitChan:
			log.Printf("got signal %s, shutting down...", sig)
			sm.shutdown()
			break loop
		case <-childChan:
			offender := sm.findTerminatedChild()
			if offender == nil {
				log.Printf("unable to find terminated child process, ignoring SIGCHLD")
				continue
			}
			log.Printf("service %s exited with code %d", offender.config.Name, *offender.exitCode)
			if offender.config.OnExit == OnExitTerminateIfError {
				if *offender.exitCode == 0 {
					log.Printf("OnExitTerminateIfError is configured for %s and there was no error, doing nothing service termination",
						offender.config.Name)
					continue
				} else {
					log.Printf("OnExitTerminateIfError is configured for %s and there was error, shutting down...",
						offender.config.Name)
					sm.shutdown()
					return *offender.exitCode
				}
			}
			if offender.config.OnExit == OnExitRestart {
				restartsLeft := offender.config.MaxRestarts - offender.restarts
				if restartsLeft > 0 {
					log.Printf("OnExitRestart is configured for %s, %d restarts left, attempting to restart...",
						offender.config.Name, restartsLeft)
					offender.restarts++
					err := offender.Start()
					if err != nil {
						log.Printf("unable to restart service %s because of error: '%s', exitting...",
							offender.config.Name, err)
						return -1
					}
				}
			}
			if offender.config.OnExit == OnExitTerminate {
				log.Printf("OnExitTerminate is configured for %s, shutting down...", offender.config.Name)
				sm.shutdown()
				return *offender.exitCode
			}
			if offender.config.OnExit == OnExitDoNothing {
				log.Printf("OnExitDoNothing is configured for %s, doing nothing about service termination",
					offender.config.Name)
			}
		}
	}
	return 0
}

func (sm *ServiceManager) shutdown() {
	for i := 0; i < len(sm.services); i++ {
		s := sm.services[i]
		if !s.stopped {
			log.Printf("sending %s to %s...\n", s.config.StopSignalStr, s.config.Name)
			err := s.process.Signal(s.config.StopSignal)
			if err != nil {
				log.Printf("unable to send termination signal to %s : %s", s.config.Name, err)
				// it's not actually stopped but we need to exclude it from wait
				s.stopped = true
			}
		}
	}
	var wg sync.WaitGroup
	var list []*Service
	var stoppedServices = make(map[*Service]bool)
	for i := 0; i < len(sm.services); i++ {
		s := sm.services[i]
		if !s.stopped {
			wg.Add(1)
			list = append(list, s)
			stoppedServices[s] = false
		}
	}
	sc := make(chan *Service, 1)
	for i := 0; i < len(list); i++ {
		s := list[i]
		go func() {
			defer wg.Done()
			log.Printf("waiting %s to finish...\n", s.config.Name)
			state, err := s.process.Wait()
			if err != nil {
				log.Printf("cannot get state for service %s: %s", s.config.Name, err)
			}
			code := state.ExitCode()
			s.exitCode = &code
			sc <- s
		}()
	}
	c1 := make(chan bool, 1)
	go func() {
		wg.Wait()
		c1 <- true
	}()

loop:
	for {
		select {
		case <-c1:
			log.Printf("successfully shutted down")
			break loop
		case ss := <-sc:
			log.Printf("successfully stopped %s", ss.config.Name)
			stoppedServices[ss] = true
		case <-time.After(time.Duration(sm.config.ShutdownTimeoutSec) * time.Second):
			var notStopped []string
			for k, v := range stoppedServices {
				if !v {
					notStopped = append(notStopped, k.config.Name)
				} else {
					k.closeOutputs()
				}
			}
			log.Printf("shutdown timeout in %d seconds because following services are still running: %s", sm.config.ShutdownTimeoutSec, strings.Join(notStopped, ","))
			break loop
		}
	}
}

func (sm *ServiceManager) findTerminatedChild() *Service {
	var status syscall.WaitStatus
	pid, err := syscall.Wait4(-1, &status, 0, nil)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(sm.services); i++ {
		p := sm.services[i]
		if p.process.Pid == pid {
			p.stopped = true
			code := status.ExitStatus()
			p.exitCode = &code
			return p
		}
	}
	return nil
}
