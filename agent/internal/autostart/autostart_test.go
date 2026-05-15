package autostart

import (
	"testing"
)

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		entry    Entry
		want     string
		wantErr  bool
	}{
		{
			name: "simple path no args",
			entry: Entry{
				ExecutablePath: `C:\tools\app.exe`,
			},
			want: `"C:\tools\app.exe"`,
		},
		{
			name: "path with args",
			entry: Entry{
				ExecutablePath: `C:\tools\app.exe`,
				Args:           []string{"--port", "8080"},
			},
			want: `"C:\tools\app.exe" --port 8080`,
		},
		{
			name: "path with spaces",
			entry: Entry{
				ExecutablePath: `C:\Program Files\My App\app.exe`,
				Args:           []string{"--config", `C:\config\my config.json`},
			},
			want: `"C:\Program Files\My App\app.exe" --config "C:\config\my config.json"`,
		},
		{
			name:    "empty executable path",
			entry:   Entry{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildCommand(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		cmd             string
		wantExe         string
		wantArgs        []string
	}{
		{`"C:\tools\app.exe"`, `C:\tools\app.exe`, nil},
		{`"C:\tools\app.exe" --port 8080`, `C:\tools\app.exe`, []string{"--port", "8080"}},
		{`"C:\Program Files\app.exe" --config "my config.json"`, `C:\Program Files\app.exe`, []string{"--config", `"my config.json"`}},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			gotExe, gotArgs := ParseCommand(tt.cmd)
			if gotExe != tt.wantExe {
				t.Errorf("ParseCommand() exe = %q, want %q", gotExe, tt.wantExe)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("ParseCommand() args = %v, want %v", gotArgs, tt.wantArgs)
				return
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("ParseCommand() arg[%d] = %q, want %q", i, gotArgs[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestRegistryKeyName(t *testing.T) {
	entry := Entry{ID: "uuid-123"}
	want := "LocalServicePanel_CustomApp_uuid-123"
	if got := entry.RegistryKeyName(); got != want {
		t.Errorf("RegistryKeyName() = %q, want %q", got, want)
	}
}

func TestServiceError(t *testing.T) {
	err := NewServiceError(ErrCodeUnsupported, "not supported", nil)
	if err.Code != ErrCodeUnsupported {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeUnsupported)
	}
	if err.Error() == "" {
		t.Error("Error() should not be empty")
	}
}
