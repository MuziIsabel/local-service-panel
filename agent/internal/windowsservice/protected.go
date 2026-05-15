// Package windowsservice provides Windows Service management.
package windowsservice

import "github.com/user/local-service-panel/agent/internal/domain"

// DefaultProtectedServices is the initial list of Windows Services
// considered critical or high-risk. These services should be protected
// from dangerous operations (stop, restart, disable).
//
// Service names may vary across Windows versions.
var DefaultProtectedServices = map[domain.ServiceName]bool{
	"WinDefend":              true, // Windows Defender
	"EventLog":               true, // Windows Event Log
	"RpcSs":                  true, // Remote Procedure Call (RPC)
	"PlugPlay":               true, // Plug and Play
	"SamSs":                  true, // Security Accounts Manager
	"LSM":                    true, // Local Session Manager
	"Winmgmt":                true, // Windows Management Instrumentation
	"DcomLaunch":             true, // DCOM Server Process Launcher
	"SecurityHealthService":  true, // Windows Security Health Service
	"mpssvc":                 true, // Windows Defender Firewall
}

// IsProtected checks whether the given service name is in the protected list.
func IsProtected(name domain.ServiceName) bool {
	return DefaultProtectedServices[name]
}

// IsHighRiskAction determines if an operation on a protected service
// should be rejected. Start actions are generally allowed.
func IsHighRiskAction(serviceName domain.ServiceName, action string) bool {
	if !IsProtected(serviceName) {
		return false
	}
	switch action {
	case "stop", "restart", "set_start_type":
		return true
	}
	return false
}
