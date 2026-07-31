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
