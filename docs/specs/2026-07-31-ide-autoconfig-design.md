# Design Doc: Autoconfiguração de IDEs (IntelliJ IDEA e VS Code)

**Data**: 2026-07-31  
**Autor**: Jackson Wendel Santos Sá & Antigravity  
**Status**: Aprovado  
**Issue Relacionada**: [#2](https://github.com/wendelmax/ezship/issues/2)

---

## 1. Visão Geral

O objetivo desta funcionalidade é fornecer o comando `ezship setup ide` para autoconfigurar a integração de containers do **Docker Engine** (seja rodando em modo **WSL2** ou **Windows Native**) em IDEs populares no Windows, inicialmente **IntelliJ IDEA** e **VS Code**.

Ao executar o comando, o `ezship` detecta automaticamente as IDEs instaladas no sistema, faz backup das configurações existentes e injeta as strings de conexão e arquivos XML/JSON necessários.

---

## 2. Arquitetura (`internal/ide`)

Definição da interface principal no pacote `internal/ide/ide.go`:

```go
package ide

type IDEConfigurator interface {
	Name() string
	IsInstalled() bool
	Configure(dockerMode string) error // dockerMode: "native" ou "wsl"
}

func DetectIDEs() []IDEConfigurator {
	// Retorna instâncias de configuradores para IDEs detectadas no sistema
}
```

---

## 3. Implementação IntelliJ IDEA (`internal/ide/intellij.go`)

- **Diretório Alvo**: `%USERPROFILE%\AppData\Roaming\JetBrains\<VersaoIntelliJ>\options\`
- **Arquivos Manipulados**:
  - `docker-tools.xml`
  - `remote-servers.xml`
- **Configuração de Conexão**:
  - **Modo Native**: Named Pipe `npipe:////./pipe/docker_engine`
  - **Modo WSL2**: `unix:///var/run/docker.sock` ou `tcp://localhost:2375`
- **Backup**: Salva cópia de segurança em `remote-servers.xml.ezship.bak` antes de qualquer modificação.

---

## 4. Implementação VS Code (`internal/ide/vscode.go`)

- **Diretório Alvo**: `%USERPROFILE%\AppData\Roaming\Code\User\settings.json`
- **Chave Atualizada**: `"docker.host"`
- **Backup**: Salva cópia de segurança em `settings.json.ezship.bak`.

---

## 5. CLI & Menu Interativo

- Subcomando: `ezship setup ide [intellij|vscode]`
- Caso executado sem argumentos (`ezship setup ide`), apresenta um menu interativo enumerando as IDEs detectadas.
