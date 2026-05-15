// Package service provides Windows Service lifecycle management for the Agent.
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	Name        = "LocalServicePanelAgent"
	DisplayName = "Local Service Panel Agent"
	Description = "Local Service Panel Agent - manages Windows services, custom apps, and startup items"
)

// RunFn is a function that runs the agent and blocks until ctx is cancelled.
type RunFn func(ctx context.Context) error

// Install registers the Agent as a Windows Service.
func Install() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}

	// Pass -service run so the agent knows to run in service mode.
	cmdLine := fmt.Sprintf(`"%s" -service run`, exePath)

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.CreateService(Name, cmdLine, mgr.Config{
		DisplayName:      DisplayName,
		Description:      Description,
		StartType:        mgr.StartAutomatic,
		DelayedAutoStart: true,
	})
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}
	s.Close()
	return nil
}

// Uninstall removes the Agent Windows Service.
func Uninstall() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(Name)
	if err != nil {
		return fmt.Errorf("open service: %w", err)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

// RunAsService runs the agent function as a Windows Service.
func RunAsService(runFn RunFn) error {
	isSvc, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("check if running as service: %w", err)
	}
	if !isSvc {
		return fmt.Errorf("not running as a Windows Service")
	}
	return svc.Run(Name, &handler{runFn: runFn})
}

type handler struct {
	runFn RunFn
}

func (h *handler) Execute(args []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.runFn(ctx)
	}()

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case c := <-requests:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				cancel()
				changes <- svc.Status{State: svc.StopPending}
				return false, 0
			default:
				continue
			}
		case err := <-errCh:
			if err != nil {
				return false, 1
			}
			return false, 0
		}
	}
}
