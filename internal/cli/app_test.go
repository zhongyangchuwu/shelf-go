package cli

import (
	"github.com/zhongyangchuwu/shelf-go/internal/app"
	"github.com/zhongyangchuwu/shelf-go/internal/jsonvault"
)

func testApp() *app.App {
	return app.New(jsonvault.Provider{})
}
