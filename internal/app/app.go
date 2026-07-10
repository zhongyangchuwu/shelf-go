package app

type App struct{}

func New() *App {
	return &App{}
}

func NewDefault() *App {
	return New()
}
