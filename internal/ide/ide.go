package ide

type IDEConfigurator interface {
	Name() string
	IsInstalled() bool
	Configure(dockerMode string) error
}

func DetectIDEs() []IDEConfigurator {
	detected := make([]IDEConfigurator, 0)
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
