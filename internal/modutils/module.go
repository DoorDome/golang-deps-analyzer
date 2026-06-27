package modutils

type Module struct {
	Name         string
	GoVersion    string
	AbsolutePath string
	RelativePath string
	Dependencies []Dependency
}

type Dependency struct {
	Path     string
	Version  string
	Indirect bool
}
