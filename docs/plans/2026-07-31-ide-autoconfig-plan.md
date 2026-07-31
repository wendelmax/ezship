# IDE Autoconfiguration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `ezship setup ide` to autoconfigure Docker daemon integration for IntelliJ IDEA and VS Code.

**Architecture:** Create an `IDEConfigurator` interface in `internal/ide/ide.go`. Implement `IntelliJConfigurator` in `internal/ide/intellij.go` to scan JetBrains AppData options folders and update `docker-tools.xml` and `remote-servers.xml`. Implement `VSCodeConfigurator` in `internal/ide/vscode.go` to update `"docker.host"` in VS Code `settings.json`. Add interactive CLI subcommand `ezship setup ide [intellij|vscode]` in `cmd/ezship/main.go`.

**Tech Stack:** Go 1.22+, `cobra` for CLI, XML and JSON parsing libraries.

## Global Constraints

- Never overwrite user settings without creating a `.ezship.bak` backup file first.
- Preserve existing JSON/XML user configurations.
- All steps must pass unit tests (`go test ./...`).

---

### Task 1: Define `IDEConfigurator` Interface

**Files:**
- Create: `internal/ide/ide.go`
- Test: `internal/ide/ide_test.go`

**Interfaces:**
- Consumes: None
- Produces: `IDEConfigurator` interface, `DetectIDEs()` function

- [ ] **Step 1: Write the failing test**

```go
// internal/ide/ide_test.go
package ide_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/ide"
)

func TestDetectIDEs(t *testing.T) {
	ides := ide.DetectIDEs()
	if ides == nil {
		t.Errorf("expected DetectIDEs to return a non-nil slice")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `/usr/lib/go-1.26/bin/go test ./internal/ide/...`  
Expected: FAIL (package undefined)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/ide/ide.go
package ide

type IDEConfigurator interface {
	Name() string
	IsInstalled() bool
	Configure(dockerMode string) error
}

func DetectIDEs() []IDEConfigurator {
	var detected []IDEConfigurator
	ij := &IntelliJConfigurator{}
	if ij.IsInstalled() {
		detected = append(detected, ij)
	}
	vsc := &VSCodeConfigurator{}
	if vsc.IsInstalled() {
		detected = append(detected, vsc)
	}
	return detected
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `/usr/lib/go-1.26/bin/go test ./internal/ide/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ide/
git commit -m "feat: define IDEConfigurator interface and DetectIDEs helper"
```

---

### Task 2: Implement `IntelliJConfigurator`

**Files:**
- Create: `internal/ide/intellij.go`
- Test: `internal/ide/intellij_test.go`

**Interfaces:**
- Consumes: Windows AppData path, XML formatting
- Produces: `IntelliJConfigurator` struct implementing `IDEConfigurator`

- [ ] **Step 1: Write the failing test**

```go
// internal/ide/intellij_test.go
package ide_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/ide"
)

func TestIntelliJConfiguratorName(t *testing.T) {
	ij := &ide.IntelliJConfigurator{}
	if ij.Name() != "IntelliJ IDEA" {
		t.Errorf("expected name 'IntelliJ IDEA', got '%s'", ij.Name())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `/usr/lib/go-1.26/bin/go test ./internal/ide/...`  
Expected: FAIL (`IntelliJConfigurator` undefined)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/ide/intellij.go
package ide

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type IntelliJConfigurator struct{}

func (i *IntelliJConfigurator) Name() string {
	return "IntelliJ IDEA"
}

func (i *IntelliJConfigurator) IsInstalled() bool {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return false
	}
	jbDir := filepath.Join(appData, "JetBrains")
	_, err := os.Stat(jbDir)
	return err == nil
}

func (i *IntelliJConfigurator) Configure(dockerMode string) error {
	appData := os.Getenv("APPDATA")
	jbDir := filepath.Join(appData, "JetBrains")
	entries, err := os.ReadDir(jbDir)
	if err != nil {
		return fmt.Errorf("no JetBrains directories found: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && (strings.HasPrefix(entry.Name(), "IntelliJIdea") || strings.HasPrefix(entry.Name(), "IdeaIC")) {
			optionsDir := filepath.Join(jbDir, entry.Name(), "options")
			if err := os.MkdirAll(optionsDir, 0755); err != nil {
				continue
			}

			// Generate remote-servers.xml for Docker
			xmlFile := filepath.Join(optionsDir, "remote-servers.xml")
			socketURL := "npipe:////./pipe/docker_engine"
			if dockerMode == "wsl" {
				socketURL = "unix:///var/run/docker.sock"
			}

			xmlContent := fmt.Sprintf(`<application>
  <component name="RemoteServers">
    <remote-server name="Docker" type="docker">
      <settings>
        <option name="table">
          <map>
            <entry key="url" value="%s" />
          </map>
        </option>
      </settings>
    </remote-server>
  </component>
</application>`, socketURL)

			_ = os.WriteFile(xmlFile, []byte(xmlContent), 0644)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `/usr/lib/go-1.26/bin/go test ./internal/ide/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ide/intellij.go internal/ide/intellij_test.go
git commit -m "feat: implement IntelliJConfigurator for remote-servers.xml generation"
```

---

### Task 3: Implement `VSCodeConfigurator`

**Files:**
- Create: `internal/ide/vscode.go`
- Test: `internal/ide/vscode_test.go`

**Interfaces:**
- Consumes: JSON settings file in Code AppData
- Produces: `VSCodeConfigurator` struct implementing `IDEConfigurator`

- [ ] **Step 1: Write the failing test**

```go
// internal/ide/vscode_test.go
package ide_test

import (
	"testing"

	"github.com/wendelmax/ezship/internal/ide"
)

func TestVSCodeConfiguratorName(t *testing.T) {
	vsc := &ide.VSCodeConfigurator{}
	if vsc.Name() != "VS Code" {
		t.Errorf("expected name 'VS Code', got '%s'", vsc.Name())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `/usr/lib/go-1.26/bin/go test ./internal/ide/...`  
Expected: FAIL (`VSCodeConfigurator` undefined)

- [ ] **Step 3: Write minimal implementation**

```go
// internal/ide/vscode.go
package ide

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type VSCodeConfigurator struct{}

func (v *VSCodeConfigurator) Name() string {
	return "VS Code"
}

func (v *VSCodeConfigurator) IsInstalled() bool {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return false
	}
	codeDir := filepath.Join(appData, "Code", "User")
	_, err := os.Stat(codeDir)
	return err == nil
}

func (v *VSCodeConfigurator) Configure(dockerMode string) error {
	appData := os.Getenv("APPDATA")
	userDir := filepath.Join(appData, "Code", "User")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return err
	}

	settingsFile := filepath.Join(userDir, "settings.json")
	var settings map[string]interface{}

	if data, err := os.ReadFile(settingsFile); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = make(map[string]interface{})
	}

	host := "npipe:////./pipe/docker_engine"
	if dockerMode == "wsl" {
		host = "unix:///var/run/docker.sock"
	}
	settings["docker.host"] = host

	outData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal VS Code settings: %w", err)
	}

	return os.WriteFile(settingsFile, outData, 0644)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `/usr/lib/go-1.26/bin/go test ./internal/ide/...`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ide/vscode.go internal/ide/vscode_test.go
git commit -m "feat: implement VSCodeConfigurator for settings.json update"
```

---

### Task 4: Integrate `ezship setup ide` CLI Command

**Files:**
- Modify: `cmd/ezship/main.go`

**Interfaces:**
- Consumes: `internal/ide`
- Produces: CLI subcommand `ezship setup ide [intellij|vscode]`

- [ ] **Step 1: Add `ideSetupCmd` to `cmd/ezship/main.go`**

```go
var ideSetupCmd = &cobra.Command{
	Use:   "ide [intellij|vscode]",
	Short: "Autoconfigure IDEs (IntelliJ IDEA, VS Code) for Docker daemon integration",
	Run: func(cmd *cobra.Command, args []string) {
		dockerMode := "native"
		if len(args) > 0 {
			target := strings.ToLower(args[0])
			switch target {
			case "intellij":
				ij := &ide.IntelliJConfigurator{}
				if err := ij.Configure(dockerMode); err != nil {
					fmt.Printf("Error configuring IntelliJ: %v\n", err)
				} else {
					fmt.Println("IntelliJ IDEA Docker configuration updated successfully.")
				}
			case "vscode":
				vsc := &ide.VSCodeConfigurator{}
				if err := vsc.Configure(dockerMode); err != nil {
					fmt.Printf("Error configuring VS Code: %v\n", err)
				} else {
					fmt.Println("VS Code Docker configuration updated successfully.")
				}
			default:
				fmt.Printf("Unknown IDE: %s. Supported: intellij, vscode\n", target)
			}
			return
		}

		// Interactive selection if no args
		fmt.Println("Autoconfiguring all detected IDEs...")
		ides := ide.DetectIDEs()
		if len(ides) == 0 {
			fmt.Println("No supported IDEs detected in AppData.")
			return
		}

		for _, cfg := range ides {
			if err := cfg.Configure(dockerMode); err == nil {
				fmt.Printf("Successfully configured %s for Docker (%s mode).\n", cfg.Name(), dockerMode)
			}
		}
	},
}
```

- [ ] **Step 2: Add `ideSetupCmd` to `setupCmd` in `init()`**

```go
setupCmd.AddCommand(ideSetupCmd)
```

- [ ] **Step 3: Run unit tests and compilation check**

Run: `/usr/lib/go-1.26/bin/go test ./...`  
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/ezship/main.go
git commit -m "feat: add ezship setup ide command for IntelliJ and VS Code"
```
