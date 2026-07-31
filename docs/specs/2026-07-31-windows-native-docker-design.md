# Design Doc: Suporte a Windows Native Docker no ezship

**Data**: 2026-07-31  
**Autor**: Jackson Wendel Santos Sá & Antigravity  
**Status**: Aprovado  
**Issue Relacionada**: [#1](https://github.com/wendelmax/ezship/issues/1)

---

## 1. Visão Geral

O objetivo desta funcionalidade é estender o **ezship** para suportar a execução e instalação do **Docker Engine (CLI + Daemon + Compose)** diretamente no **Windows Host (Modo Nativo)**, eliminando a dependência obrigatória do WSL2 para usuários que preferem não utilizar máquinas virtuais ou a camada Linux.

Ao mesmo tempo, o `ezship` manterá seu suporte completo ao **WSL2** como modo padrão/multi-engine (Docker, Podman, nerdctl, k3s, k3d).

---

## 2. Arquitetura: Provider Interface

Para manter o código modular, testável e expansível, introduziremos o padrão **Provider Interface** no pacote `internal/provider`.

### 2.1 Estrutura de Tipos (`internal/provider/provider.go`)

```go
package provider

type EngineStatus struct {
	Name    string
	Running bool
	Version string
	Mode    string // "WSL2" ou "Native"
}

type EngineProvider interface {
	Name() string // "wsl" ou "native"
	IsEngineSupported(engine string) bool
	InstallEngine(engine string) error
	EnsureEngineRunning(engine string) error
	StopEngine(engine string) error
	GetStatus(engine string) EngineStatus
	RunProxyCommand(engine string, args []string) error
}
```

### 2.2 Implementações

1. **`WSLProvider` (`internal/provider/wsl.go`)**: Encapsula as operações existentes de importação do Ubuntu Minimal e execução de daemons no WSL2.
2. **`NativeProvider` (`internal/provider/native.go`)**: Gerencia o download e execução nativos dos binários Windows.

---

## 3. Funcionamento do `NativeProvider` (Windows Host)

### 3.1 Estrutura de Arquivos no Windows Host
- **Binários**: `%USERPROFILE%\.ezship\bin\` (`docker.exe`, `dockerd.exe`, `docker-compose.exe`).
- **Configuração do ezship**: `%USERPROFILE%\.ezship\config.json`.
- **Logs do Daemon**: `%USERPROFILE%\.ezship\logs\dockerd.log`.

### 3.2 Download Automático
- Binários do Docker Engine (`docker.exe`, `dockerd.exe`) são baixados dos archives estáticos oficiais da Docker Inc (`https://download.docker.com/win/static/stable/x86_64/`).
- Binário do Docker Compose (`docker-compose.exe`) é baixado dos Releases oficiais no GitHub.

### 3.3 Daemon & Socket Nativo
- O `dockerd.exe` é executado no Windows conectado ao Named Pipe nativo: `npipe:////./pipe/docker_engine`.
- O `PATH` do usuário no Windows é atualizado para incluir `%USERPROFILE%\.ezship\bin`, permitindo rodar `docker` e `docker-compose` diretamente de qualquer terminal (PowerShell, CMD, Bash).

---

## 4. CLI e TUI Integration

### 4.1 CLI Comandos
- `ezship setup docker`: Instala via **WSL2** por padrão (compatibilidade retroativa).
- `ezship setup docker --native`: Instala no modo **Windows Native**.
- `ezship status`: Exibe o estado de todas as engines e o modo de execução:
  ```text
  ENGINE     STATUS       VERSION        MODE
  ---------------------------------------------------
  docker     Running      27.5.1         Native
  podman     Stopped      5.0.0          WSL2
  ```

### 4.2 TUI Dashboard (`internal/tui`)
Ao selecionar a opção **Setup** para o Docker na TUI, um modal interativo permitirá a escolha entre:
- `[ ] WSL2 (Ubuntu Minimal)`
- `[ ] Windows Native (Sem WSL2)`

---

## 5. Manutenção e Limpeza

- `ezship prune`: Executa a limpeza em ambos os providers se estiverem rodando.
- `ezship vacuum`: Notifica o usuário de que a operação se aplica apenas aos discos `.vhdx` do WSL2.
- `ezship reset`: Adiciona a limpeza dos binários nativos em `%USERPROFILE%\.ezship\bin` além da remoção da distro WSL2.
