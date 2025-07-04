package app

type App struct {
	Name    string
	Version string
}

func NewApp(name string, version string) *App {
	return &App{
		Name:    name,
		Version: version,
	}
}

func (a *App) Start() {
}