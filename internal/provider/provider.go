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
