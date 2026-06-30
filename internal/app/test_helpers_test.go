package app

import "github.com/zhongyangchuwu/shelf-go/internal/jsonvault"

func testApp() *App {
	return New(jsonvault.Provider{})
}
