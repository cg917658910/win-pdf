package main

import (
	"context"
	"fmt"
	"os"
	"reggen/license"
	"strings"

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
	//a.initConfig()
}
func (a *App) initConfig() {
	priPath := "./config/server_private.pem"
	if _, err := os.Stat(priPath); err != nil {
		rt.LogErrorf(a.ctx, "私钥文件 %s 未找�?err %s", priPath, err)
	}
	if err := license.LoadPrivateKeyFromFile(priPath); err != nil {
		rt.LogErrorf(a.ctx, "加载私钥失败: %v", err)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GenerateRegCode(machineCode string, expiryUnix int64) (string, error) {
	return license.GenerateRegCode(machineCode, expiryUnix)
}

func (a *App) MessageDialog(title, message, dialogType string) (string, error) {
	res, err := rt.MessageDialog(a.ctx, rt.MessageDialogOptions{
		Title:   title,
		Message: message,
		Type:    mapDialogType(dialogType),
	})
	if err != nil {
		rt.LogPrintf(a.ctx, "MessageDialog error: %v", err)
		return "", err
	}
	return res, nil
}

func mapDialogType(t string) rt.DialogType {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "error", "err":
		return rt.ErrorDialog
	case "warning", "warn":
		return rt.WarningDialog
	case "question", "ask", "confirm":
		return rt.QuestionDialog
	default:
		return rt.InfoDialog
	}
}
