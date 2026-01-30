package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cg917658910/win-pdf/internal/engine"
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

// 设置有效期
func (a *App) SetExpiry(opts engine.Options) string {
	rt.LogPrintf(a.ctx, "设置有效期: %+v", opts)
	// do something
	// 检查opts.Files和opts.OutputDir
	// Todo:
	// 1.Files最少1个，最多上传10个，;这个隔开的
	filesNum := strings.Count(opts.Files, ";") + 1
	if filesNum < 1 || filesNum > 10 {
		return fmt.Sprintf("错误：请选择1到10个文件，当前选择了%d个文件", filesNum)
	}
	// 2.OutputDir不能为空
	if strings.TrimSpace(opts.OutputDir) == "" {
		return fmt.Sprintf("错误：请选择输出目录")
	}
	// 3.有效期区间最短有1分钟，最高10年
	if opts.StartTime.IsZero() || opts.EndTime.IsZero() {
		return fmt.Sprintf("错误：请设置有效的开始时间和结束时间")
	}
	if opts.EndTime.Sub(opts.StartTime) < 1*time.Minute {
		return fmt.Sprintf("错误：有效期区间最短为1分钟")
	}
	if opts.EndTime.Sub(opts.StartTime) > 10*365*24*time.Hour {
		return fmt.Sprintf("错误：有效期区间最长为10年")
	}
	err := engine.RunBatch(opts)
	if err != nil {
		return fmt.Sprintf("错误：处理文件时出错：%v", err)
	}
	return fmt.Sprintf("设置成功")
}

// OpenDirectoryDialog 打开原生选择目录对话框并返回用户选择的目录路径
func (a *App) OpenDirectoryDialog() (string, error) {
	// 使用 Wails runtime 提供的 OpenDirectoryDialog
	path, err := rt.OpenDirectoryDialog(a.ctx, rt.OpenDialogOptions{Title: "选择保存目录"})
	if err != nil {
		rt.LogPrintf(a.ctx, "OpenDirectoryDialog error: %v", err)
		return "", err
	}
	rt.LogPrintf(a.ctx, "用户选择的目录: %s", path)
	return path, nil
}

// OpenMultipleFilesDialog 打开原生多文件选择对话框并返回所选文件路径列表
func (a *App) OpenMultipleFilesDialog() ([]string, error) {
	paths, err := rt.OpenMultipleFilesDialog(a.ctx, rt.OpenDialogOptions{Title: "选择文件"})
	if err != nil {
		rt.LogPrintf(a.ctx, "OpenMultipleFilesDialog error: %v", err)
		return nil, err
	}
	return paths, nil
}

// MessageDialog 使用 Wails 原生对话框显示消息并返回用户选择结果
func (a *App) MessageDialog(title, message string) (string, error) {
	// 可选设置类型/按钮等，使用默认选项显示信息
	res, err := rt.MessageDialog(a.ctx, rt.MessageDialogOptions{Title: title, Message: message})
	if err != nil {
		rt.LogPrintf(a.ctx, "MessageDialog error: %v", err)
		return "", err
	}
	return res, nil
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
