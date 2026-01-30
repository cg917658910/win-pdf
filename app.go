package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	rt "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// window menu
func NewAppMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	FileMenu := appMenu.AddSubmenu("文件")
	FileMenu.AddText("选择文件", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		// do something
	})
	FileMenu.AddSeparator()
	FileMenu.AddText("退出", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		// `rt` is an alias of "github.com/wailsapp/wails/v2/pkg/runtime" to prevent collision with standard package
		rt.Quit(app.ctx)
	})
	// 注册menu
	registerMenu := appMenu.AddSubmenu("注册")
	registerMenu.AddText("注册", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		// do something
	})
	registerMenu.AddText("在线购买", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		// do something
	})
	registerMenu.AddText("郑重声明", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		// do something
	})
	registerMenu.AddText("联系我们", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		// do something
	})

	return appMenu
}
