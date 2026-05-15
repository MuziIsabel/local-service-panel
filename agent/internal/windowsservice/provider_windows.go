//go:build windows

package windowsservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/user/local-service-panel/agent/internal/domain"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// windowsProvider implements Provider on Windows via SCM.
type windowsProvider struct{}

// NewProvider creates a new Windows Service provider.
// On non-Windows platforms, this returns a stub that returns ErrCodeUnsupported.
func NewProvider() Provider {
	return &windowsProvider{}
}

func (p *windowsProvider) List(ctx context.Context, filter Filter) ([]domain.Service, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, NewServiceError(ErrCodeQueryFailed, "connect to SCM", err)
	}
	defer m.Disconnect()

	names, err := m.ListServices()
	if err != nil {
		return nil, NewServiceError(ErrCodeQueryFailed, "list services", err)
	}

	var services []domain.Service
	for _, name := range names {
		svc, err := queryService(m, name)
		if err != nil {
			// Skip services that fail to query (e.g. access denied)
			continue
		}

		// Apply filters
		if !matchesFilter(svc, filter) {
			continue
		}

		services = append(services, *svc)
	}

	return services, nil
}

func (p *windowsProvider) Get(ctx context.Context, serviceName string) (*domain.Service, error) {
	m, err := mgr.Connect()
	if err != nil {
		return nil, NewServiceError(ErrCodeQueryFailed, "connect to SCM", err)
	}
	defer m.Disconnect()

	svc, err := queryService(m, serviceName)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (p *windowsProvider) Start(ctx context.Context, serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return NewServiceError(ErrCodeStartFailed, "connect to SCM", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return NewServiceError(ErrCodeNotFound, fmt.Sprintf("service %q not found", serviceName), err)
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		return NewServiceError(ErrCodeStartFailed, fmt.Sprintf("start service %q", serviceName), err)
	}
	return nil
}

func (p *windowsProvider) Stop(ctx context.Context, serviceName string) error {
	if IsHighRiskAction(domain.ServiceName(serviceName), "stop") {
		return NewServiceError(ErrCodeProtected,
			fmt.Sprintf("service %q is protected and cannot be stopped", serviceName), nil)
	}

	m, err := mgr.Connect()
	if err != nil {
		return NewServiceError(ErrCodeStopFailed, "connect to SCM", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return NewServiceError(ErrCodeNotFound, fmt.Sprintf("service %q not found", serviceName), err)
	}
	defer s.Close()

	if _, err := s.Control(svc.Stop); err != nil {
		return NewServiceError(ErrCodeStopFailed, fmt.Sprintf("stop service %q", serviceName), err)
	}
	return nil
}

func (p *windowsProvider) Restart(ctx context.Context, serviceName string) error {
	if IsHighRiskAction(domain.ServiceName(serviceName), "restart") {
		return NewServiceError(ErrCodeProtected,
			fmt.Sprintf("service %q is protected and cannot be restarted", serviceName), nil)
	}

	// Stop first
	if err := p.Stop(ctx, serviceName); err != nil {
		// If already stopped or protected, skip stop error for restart
		var svcErr *ServiceError
		if errors.As(err, &svcErr) {
			if svcErr.Code == ErrCodeProtected {
				return err // protected service should not be restarted
			}
			// Other errors (not found, already stopped) — continue to start
		}
	}

	// Then start
	if err := p.Start(ctx, serviceName); err != nil {
		return NewServiceError(ErrCodeRestartFailed, fmt.Sprintf("restart service %q", serviceName), err)
	}
	return nil
}

func (p *windowsProvider) SetStartType(ctx context.Context, serviceName string, startType domain.StartType) error {
	if IsHighRiskAction(domain.ServiceName(serviceName), "set_start_type") {
		return NewServiceError(ErrCodeProtected,
			fmt.Sprintf("service %q is protected and cannot change start type", serviceName), nil)
	}

	m, err := mgr.Connect()
	if err != nil {
		return NewServiceError(ErrCodeStartTypeFailed, "connect to SCM", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return NewServiceError(ErrCodeNotFound, fmt.Sprintf("service %q not found", serviceName), err)
	}
	defer s.Close()

	cfg, err := s.Config()
	if err != nil {
		return NewServiceError(ErrCodeQueryFailed, fmt.Sprintf("get config for service %q", serviceName), err)
	}

	cfg.StartType = toSCMStartType(startType)
	cfg.DelayedAutoStart = (startType == domain.StartTypeAutomaticDelayed)

	// Use ChangeServiceConfigW directly (ModifyConfig not available in this version)
	if err := windows.ChangeServiceConfig(s.Handle, windows.SERVICE_NO_CHANGE, cfg.StartType,
		windows.SERVICE_NO_CHANGE, nil, nil, nil, nil, nil, nil, nil); err != nil {
		return NewServiceError(ErrCodeStartTypeFailed, fmt.Sprintf("change start type for service %q", serviceName), err)
	}

	// Set delayed auto start if needed
	if cfg.DelayedAutoStart {
		// SERVICE_DELAYED_AUTO_START_INFO struct (defined manually as not in x/sys yet)
		type delayedAutoStartInfo struct {
			DelayedAutoStart int32
		}
		info := delayedAutoStartInfo{DelayedAutoStart: 1}
		if err := windows.ChangeServiceConfig2(s.Handle, windows.SERVICE_CONFIG_DELAYED_AUTO_START_INFO,
			(*byte)(unsafe.Pointer(&info))); err != nil {
			return NewServiceError(ErrCodeStartTypeFailed, fmt.Sprintf("set delayed auto start for service %q", serviceName), err)
		}
	}
	return nil
}

// queryService opens a service by name and returns its domain.Service representation.
func queryService(m *mgr.Mgr, name string) (*domain.Service, error) {
	s, err := m.OpenService(name)
	if err != nil {
		return nil, NewServiceError(ErrCodeNotFound, fmt.Sprintf("open service %q", name), err)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return nil, NewServiceError(ErrCodeQueryFailed, fmt.Sprintf("query service %q", name), err)
	}

	cfg, err := s.Config()
	if err != nil {
		return nil, NewServiceError(ErrCodeQueryFailed, fmt.Sprintf("get config for service %q", name), err)
	}

	return &domain.Service{
		ServiceName:         domain.ServiceName(name),
		DisplayName:         cfg.DisplayName,
		Description:         cfg.Description,
		Status:              fromSVCState(status.State),
		StartType:           fromSCMStartType(cfg.StartType, cfg.DelayedAutoStart),
		ExecutablePath:      cfg.BinaryPathName,
		PID:                 status.ProcessId,
		CanStop:             status.Accepts&svc.AcceptStop != 0,
		CanPauseAndContinue: status.Accepts&svc.AcceptPauseAndContinue != 0,
	}, nil
}

// matchesFilter checks if a service passes the given filter criteria.
func matchesFilter(svc *domain.Service, filter Filter) bool {
	if filter.Keyword != "" {
		keyword := strings.ToLower(filter.Keyword)
		if !strings.Contains(strings.ToLower(string(svc.ServiceName)), keyword) &&
			!strings.Contains(strings.ToLower(svc.DisplayName), keyword) {
			return false
		}
	}
	if filter.Status != "" && svc.Status != filter.Status {
		return false
	}
	if filter.StartType != "" && svc.StartType != filter.StartType {
		return false
	}
	if !filter.IncludeProtected && IsProtected(svc.ServiceName) {
		return false
	}
	return true
}

// fromSVCState converts svc.State to domain.ServiceStatus.
func fromSVCState(state svc.State) domain.ServiceStatus {
	switch state {
	case svc.Stopped:
		return domain.ServiceStatusStopped
	case svc.StartPending:
		return domain.ServiceStatusStartPending
	case svc.StopPending:
		return domain.ServiceStatusStopPending
	case svc.Running:
		return domain.ServiceStatusRunning
	case svc.ContinuePending:
		return domain.ServiceStatusContinuePending
	case svc.PausePending:
		return domain.ServiceStatusPausePending
	case svc.Paused:
		return domain.ServiceStatusPaused
	default:
		return domain.ServiceStatusUnknown
	}
}

// fromSCMStartType converts SCM start type + delayed flag to domain.StartType.
func fromSCMStartType(startType uint32, delayed bool) domain.StartType {
	switch startType {
	case mgr.StartDisabled:
		return domain.StartTypeDisabled
	case mgr.StartManual:
		return domain.StartTypeManual
	case mgr.StartAutomatic:
		if delayed {
			return domain.StartTypeAutomaticDelayed
		}
		return domain.StartTypeAutomatic
	default:
		return domain.StartTypeUnknown
	}
}

// toSCMStartType converts domain.StartType to SCM start type value.
func toSCMStartType(st domain.StartType) uint32 {
	switch st {
	case domain.StartTypeDisabled:
		return mgr.StartDisabled
	case domain.StartTypeManual:
		return mgr.StartManual
	case domain.StartTypeAutomatic, domain.StartTypeAutomaticDelayed:
		return mgr.StartAutomatic
	default:
		return mgr.StartAutomatic
	}
}

// AsServiceError unwraps an error chain to find a *ServiceError.
func AsServiceError(err error) *ServiceError {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		return svcErr
	}
	return nil
}
