# Windows Native Docker Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Windows Native Docker support to `ezship` alongside WSL2 support using a clean `EngineProvider` interface abstraction.

**Architecture:** Create an `EngineProvider` interface in `internal/provider/provider.go`. Implement `WSLProvider` in `internal/provider/wsl.go` wrapping existing WSL2 logic and `NativeProvider` in `internal/provider/native.go` to manage native Windows Docker binaries (`docker.exe`, `dockerd.exe`, `docker-compose.exe`) downloading to `%USERPROFILE%\.ezship\bin` and communicating via Named Pipe `npipe:////./pipe/docker_engine`.

**Tech Stack:** Go 1.22+, `cobra` for CLI, `bubbletea`/`lipgloss` for TUI, Windows API /exec for process management.

## Global Constraints

- Preserve 100% backward compatibility for existing WSL2 workflows.
- Store native binaries in `%USERPROFILE%\.ezship\bin`.
- All steps must pass unit tests (`go test ./...`).

---

### Task 1: Define `EngineProvider` Interface

**Files:**
- Create: `internal/provider/provider.go`
- Test: `internal/provider/provider_test.go`

**Interfaces:**
- Consumes: None
- Produces: `EngineStatus`, `EngineProvider` interface

- [ ] **Step 1: Write the failing test**

```go
// internal/provider/provider_test.go
package provider

import "testing"

func TestEngineStatusStruct(t *testing.T) {
	status := EngineStatus{
		Name:    "docker",
		Running: true,
		Version: "27.5.1",
		Mode:    "Native",
	}
	if status.Mode != "Native" {
		t.Errorf("expected Mode Native, got %s", status.Mode)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/provider/...`  
Expected: FAIL (package/type missing)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/provider/provider.go
package provider

type EngineStatus struct {
	Name    string
	Running bool
	Version string
	Mode    string // "WSL2" or "Native"
}

type EngineProvider interface {
	Name() string
	IsEngineSupported(engine string) bool
	InstallEngine(engine string) error
	EnsureEngineRunning(engine string) error
	StopEngine(engine string) error
	GetStatus(engine string) EngineStatus
	RunProxyCommand(engine string, args []string) error
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/provider/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/provider/
git commit -m "feat: define EngineProvider interface and EngineStatus type"
```

---

### Task 2: Implement `WSLProvider`

**Files:**
- Create: `internal/provider/wsl.go`
- Test: `internal/provider/wsl_test.go`

**Interfaces:**
- Consumes: `internal/wsl` package functions
- Produces: `WSLProvider` struct implementing `EngineProvider`

- [ ] **Step 1: Write the failing test**

```go
// internal/provider/wsl_test.go
package provider

import "testing"

func TestWSLProviderName(t *testing.T) {
	p := &WSLProvider{}
	if p.Name() != "wsl" {
		t.Errorf("expected name 'wsl', got '%s'", p.Name())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/provider/...`  
Expected: FAIL (`WSLProvider` undefined)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/provider/wsl.go
package provider

import (
	"github.com/wendelmax/ezship/internal/wsl"
)

type WSLProvider struct{}

func (w *WSLProvider) Name() string {
	return "wsl"
}

func (w *WSLProvider) IsEngineSupported(engine string) bool {
	switch engine {
	case "docker", "podman", "k3s", "nerdctl", "k3d":
		return true
	default:
		return false
	}
}

func (w *WSLProvider) InstallEngine(engine string) error {
	return wsl.InstallEngine(engine)
}

func (w *WSLProvider) EnsureEngineRunning(engine string) error {
	return wsl.EnsureEngineRunning(engine)
}

func (w *WSLProvider) StopEngine(engine string) error {
	return wsl.StopEngine(engine)
}

func (w *WSLProvider) GetStatus(engine string) EngineStatus {
	st := wsl.GetEngineStatus(engine)
	return EngineStatus{
		Name:    st.Name,
		Running: st.Running,
		Version: st.Version,
		Mode:    "WSL2",
	}
}

func (w *WSLProvider) RunProxyCommand(engine string, args []string) error {
	return wsl.RunProxyCommand(engine, args)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/provider/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/provider/wsl.go internal/provider/wsl_test.go
git commit -m "feat: implement WSLProvider wrapping internal/wsl"
```

---

### Task 3: Implement `NativeProvider` for Windows Host

**Files:**
- Create: `internal/provider/native.go`
- Test: `internal/provider/native_test.go`

**Interfaces:**
- Consumes: Windows process execution, HTTP downloads
- Produces: `NativeProvider` struct implementing `EngineProvider`

- [ ] **Step 1: Write the failing test**

```go
// internal/provider/native_test.go
package provider

import "testing"

func TestNativeProviderSupport(t *testing.T) {
	p := &NativeProvider{}
	if !p.IsEngineSupported("docker") {
		t.Errorf("expected docker to be supported natively")
	}
	if p.IsEngineSupported("podman") {
		t.Errorf("expected podman not to be supported natively in ezship")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/provider/...`  
Expected: FAIL (`NativeProvider` undefined)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/provider/native.go
package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NativeProvider struct{}

func (n *NativeProvider) Name() string {
	return "native"
}

func (n *NativeProvider) IsEngineSupported(engine string) bool {
	return strings.ToLower(engine) == "docker"
}

func (n *NativeProvider) InstallEngine(engine string) error {
	if !n.IsEngineSupported(engine) {
		return fmt.Errorf("engine %s is not supported natively on Windows", engine)
	}

	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home: %w", err)
	}

	binDir := filepath.Join(userHome, ".ezship", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	fmt.Printf("Native Docker installation path: %s\n", binDir)
	return nil
}

func (n *NativeProvider) EnsureEngineRunning(engine string) error {
	// Check if dockerd is running natively
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq dockerd.exe")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "dockerd.exe") {
		return nil
	}
	return nil
}

func (n *NativeProvider) StopEngine(engine string) error {
	cmd := exec.Command("taskkill", "/F", "/IM", "dockerd.exe")
	return cmd.Run()
}

func (n *NativeProvider) GetStatus(engine string) EngineStatus {
	running := false
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq dockerd.exe")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "dockerd.exe") {
		running = true
	}

	version := "Unknown"
	if running {
		vCmd := exec.Command("docker", "--version")
		if vOut, vErr := vCmd.Output(); vErr == nil {
			version = strings.TrimSpace(string(vOut))
		}
	}

	return EngineStatus{
		Name:    "docker",
		Running: running,
		Version: version,
		Mode:    "Native",
	}
}

func (n *NativeProvider) RunProxyCommand(engine string, args []string) error {
	cmd := exec.Command(engine, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/provider/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/provider/native.go internal/provider/native_test.go
git commit -m "feat: implement NativeProvider skeleton for Windows Host"
```

---

### Task 4: Integrate `--native` Flag & Provider Selection in CLI

**Files:**
- Modify: `cmd/ezship/main.go`
- Test: Build and test CLI `--native` flag parsing

**Interfaces:**
- Consumes: `internal/provider`
- Produces: CLI `--native` flag support in `ezship setup docker --native`

- [ ] **Step 1: Modify `cmd/ezship/main.go` to add `--native` flag**

```go
var nativeFlag bool

func init() {
	setupCmd.Flags().BoolVarP(&nativeFlag, "native", "n", false, "Install/setup engine natively on Windows (without WSL2)")
	// ... rest of init
}
```

- [ ] **Step 2: Update `setupCmd` handler**

```go
var setupCmd = &cobra.Command{
	Use:   "setup [engine]",
	Short: "Setup the ezship environment and optionally install an engine",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			engine := strings.ToLower(args[0])
			var p provider.EngineProvider
			if nativeFlag {
				p = &provider.NativeProvider{}
			} else {
				p = &provider.WSLProvider{}
			}
			if err := p.InstallEngine(engine); err != nil {
				fmt.Printf("Error installing %s: %v\n", engine, err)
				os.Exit(1)
			}
		}
	},
}
```

- [ ] **Step 3: Run `go build` to verify compilation**

Run: `go build -o bin/ezship.exe ./cmd/ezship`  
Expected: Build succeeds with 0 errors.

- [ ] **Step 4: Commit**

```bash
git add cmd/ezship/main.go
git commit -m "feat: add --native flag to ezship setup CLI"
```
