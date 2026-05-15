//go:build !windows

package autostart

// stubProvider returns ErrCodeUnsupported on non-Windows platforms.
type stubProvider struct{}

// NewProvider creates a new autostart Provider.
func NewProvider() Provider {
	return &stubProvider{}
}

func (p *stubProvider) Enable(entry Entry) error {
	return NewServiceError(ErrCodeUnsupported, "autostart management is not supported on this platform", nil)
}

func (p *stubProvider) Disable(entry Entry) error {
	return NewServiceError(ErrCodeUnsupported, "autostart management is not supported on this platform", nil)
}

func (p *stubProvider) IsEnabled(entry Entry) (bool, error) {
	return false, NewServiceError(ErrCodeUnsupported, "autostart management is not supported on this platform", nil)
}
