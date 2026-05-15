//go:build !windows

package windowsservice

import (
	"context"

	"github.com/user/local-service-panel/agent/internal/domain"
)

// stubProvider returns ErrCodeUnsupported on non-Windows platforms.
type stubProvider struct{}

// NewProvider creates a new Windows Service provider.
// On non-Windows platforms, all operations return ErrCodeUnsupported.
func NewProvider() Provider {
	return &stubProvider{}
}

func (p *stubProvider) List(ctx context.Context, filter Filter) ([]domain.Service, error) {
	return nil, NewServiceError(ErrCodeUnsupported, "Windows Service management is not supported on this platform", nil)
}

func (p *stubProvider) Get(ctx context.Context, serviceName string) (*domain.Service, error) {
	return nil, NewServiceError(ErrCodeUnsupported, "Windows Service management is not supported on this platform", nil)
}

func (p *stubProvider) Start(ctx context.Context, serviceName string) error {
	return NewServiceError(ErrCodeUnsupported, "Windows Service management is not supported on this platform", nil)
}

func (p *stubProvider) Stop(ctx context.Context, serviceName string) error {
	return NewServiceError(ErrCodeUnsupported, "Windows Service management is not supported on this platform", nil)
}

func (p *stubProvider) Restart(ctx context.Context, serviceName string) error {
	return NewServiceError(ErrCodeUnsupported, "Windows Service management is not supported on this platform", nil)
}

func (p *stubProvider) SetStartType(ctx context.Context, serviceName string, startType domain.StartType) error {
	return NewServiceError(ErrCodeUnsupported, "Windows Service management is not supported on this platform", nil)
}
