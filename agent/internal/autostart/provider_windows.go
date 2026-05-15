//go:build windows

package autostart

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const (
	registryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
)

// windowsProvider implements Provider on Windows via HKCU Run registry key.
type windowsProvider struct{}

// NewProvider creates a new autostart Provider.
func NewProvider() Provider {
	return &windowsProvider{}
}

func (p *windowsProvider) Enable(entry Entry) error {
	cmd, err := BuildCommand(entry)
	if err != nil {
		return err
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return NewServiceError(ErrCodeRegistryWriteFailed,
			fmt.Sprintf("open Run key: %v", err), err)
	}
	defer k.Close()

	if err := k.SetStringValue(entry.RegistryKeyName(), cmd); err != nil {
		return NewServiceError(ErrCodeRegistryWriteFailed,
			fmt.Sprintf("set value %q: %v", entry.RegistryKeyName(), err), err)
	}
	return nil
}

func (p *windowsProvider) Disable(entry Entry) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return NewServiceError(ErrCodeRegistryDeleteFailed,
			fmt.Sprintf("open Run key: %v", err), err)
	}
	defer k.Close()

	if err := k.DeleteValue(entry.RegistryKeyName()); err != nil {
		return NewServiceError(ErrCodeRegistryDeleteFailed,
			fmt.Sprintf("delete value %q: %v", entry.RegistryKeyName(), err), err)
	}
	return nil
}

func (p *windowsProvider) IsEnabled(entry Entry) (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return false, NewServiceError(ErrCodeRegistryOpenFailed,
			fmt.Sprintf("open Run key: %v", err), err)
	}
	defer k.Close()

	val, _, err := k.GetStringValue(entry.RegistryKeyName())
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, NewServiceError(ErrCodeRegistryOpenFailed,
			fmt.Sprintf("read value %q: %v", entry.RegistryKeyName(), err), err)
	}
	return val != "", nil
}
