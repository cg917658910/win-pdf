package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cg917658910/win-pdf/internal/engine/v2"
	"github.com/cg917658910/win-pdf/internal/license"
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
	a.initConfig()
}
func (a *App) initConfig() {
	pubPath := "./confg/server_public.pem"
	if _, err := os.Stat(pubPath); err != nil {
		rt.LogErrorf(a.ctx, "公钥文件 %s 未找到 err %s", pubPath, err)
	}
	if err := license.LoadPublicKeyFromFile(pubPath); err != nil {
		rt.LogErrorf(a.ctx, "加载公钥失败: %v", err)
	}
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
	// 至少6位
	if strings.TrimSpace(opts.UserPassword) != "" && len(opts.UserPassword) < 6 {
		return fmt.Sprintf("错误：用户密码长度至少6位")
	}
	// 未注册用户只能处理最多1个文件
	isActivated, _, err := license.IsActivated()
	if err != nil {
		rt.LogPrintf(a.ctx, "检查注册状态失败: %v", err)
	}
	if !isActivated && filesNum > 1 {
		return fmt.Sprintf("错误：未注册用户只能处理1个文件，请注册后使用更多功能")
	}
	err = engine.RunBatch(opts)
	if err != nil {
		return fmt.Sprintf("错误：处理文件时出错：%v", err)
	}

	return fmt.Sprintf("设置成功")
}

// IsRegistered 返回当前是否已注册（调用 internal/auth）
func (a *App) IsRegistered() bool {
	isActivated, _, err := license.IsActivated()
	if err != nil {
		rt.LogPrintf(a.ctx, "IsRegistered error: %v", err)
		return false
	}
	return isActivated
}

// GetMachineCode 返回当前机器码
func (a *App) GetMachineCode() (string, error) {
	return license.GetMachineCode()
}

// Register 尝试使用注册码注册应用
func (a *App) Register(code string) (string, error) {
	if err := license.ActivateWithRegCode(code); err != nil {
		rt.LogPrintf(a.ctx, "Register error: %v", err)
		return "", errors.New("注册失败")
	}
	// 发送事件，通知前端更新注册状态
	rt.EventsEmit(a.ctx, "user:registered")

	return "注册成功", nil
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
	paths, err := rt.OpenMultipleFilesDialog(a.ctx, rt.OpenDialogOptions{Title: "选择文件", Filters: []rt.FileFilter{
		{DisplayName: "PDF 文件", Pattern: "*.pdf"},
	}})
	if err != nil {
		rt.LogPrintf(a.ctx, "OpenMultipleFilesDialog error: %v", err)
		return nil, err
	}
	// 未注册用户只能选择最多1个文件&&单个文件大小不能超过500kb
	isActivated, _, err := license.IsActivated()
	if err != nil {
		rt.LogPrintf(a.ctx, "检查注册状态失败: %v", err)
	}
	if !isActivated {
		if len(paths) > 1 {
			return nil, errors.New("未注册用户只能选择1个文件")
		}
		for _, p := range paths {
			info, err := os.Stat(p)
			if err != nil {
				rt.LogPrintf(a.ctx, "获取文件信息失败: %v", err)
				return nil, errors.New("获取文件信息失败")
			}
			if info.Size() > 500*1024 {
				return nil, errors.New("未注册用户单个文件大小不能超过500KB")
			}
		}
	}
	return paths, nil
}

// MessageDialog 使用 Wails 原生对话框显示消息并返回用户选择结果
func (a *App) MessageDialog(title, message string) (string, error) {
	res, err := rt.MessageDialog(a.ctx, rt.MessageDialogOptions{Title: title, Message: message})
	if err != nil {
		rt.LogPrintf(a.ctx, "MessageDialog error: %v", err)
		return "", err
	}
	return res, nil
}
func (a *App) GetTitleWithRegStatus() string {
	title := "易诚无忧PDF文档有效期设置工具"
	isActivated, _, err1 := license.IsActivated()
	if err1 != nil {
		rt.LogPrintf(a.ctx, "检查注册状态失败: %v", err1)
	}
	if isActivated {
		title += "(已注册)"
	} else {
		title += "(未注册)"
	}
	return title
}

// window menu
func NewAppMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	FileMenu := appMenu.AddSubmenu("文件")
	FileMenu.AddText("选择文件", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		//app.OpenMultipleFilesDialog()
	})
	FileMenu.AddSeparator()
	FileMenu.AddText("退出", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		// `rt` is an alias of "github.com/wailsapp/wails/v2/pkg/runtime" to prevent collision with standard package
		rt.Quit(app.ctx)
	})
	// 注册menu
	registerMenu := appMenu.AddSubmenu("注册")
	registerMenu.AddText("注册", nil, func(_ *menu.CallbackData) {
		// 判断是否注册
		isActivated, _, err := license.IsActivated()
		if err != nil {
			rt.LogPrintf(app.ctx, "检查注册状态失败: %v", err)
		}
		if isActivated {
			// 提示已经注册
			rt.MessageDialog(app.ctx, rt.MessageDialogOptions{
				Title:   "注册信息",
				Message: "软件已注册，感谢您的支持！",
			})
			return
		}
		// 触发前端事件，前端负责弹出注册输入框
		rt.EventsEmit(app.ctx, "menu:register")
	})
	registerMenu.AddText("在线购买", nil, func(_ *menu.CallbackData) {
		// do something
		// 打开购买页面
		rt.BrowserOpenURL(app.ctx, "https://baidu.com")
	})
	registerMenu.AddText("郑重声明", nil, func(_ *menu.CallbackData) {
		// do something
		// 显示：本软件已在中华人民共和国国家版权局登记，未经许可，禁止任何单位和个人进行售卖、传播，违者必究！
		rt.MessageDialog(app.ctx, rt.MessageDialogOptions{
			Title:   "郑重声明",
			Message: "本软件已在中华人民共和国国家版权局登记，未经许可，禁止任何单位和个人进行售卖、传播，违者必究！",
		})
	})
	registerMenu.AddText("联系我们", nil, func(_ *menu.CallbackData) {
		// do something
		//显示邮箱地址
		rt.MessageDialog(app.ctx, rt.MessageDialogOptions{
			Title:   "联系我们",
			Message: "如有任何问题或建议，请联系邮箱：",
		})
	})
	registerMenu.AddText("注销", nil, func(_ *menu.CallbackData) {
		license.Deactivate()
		// 提示注销成功
		rt.MessageDialog(app.ctx, rt.MessageDialogOptions{
			Title:   "注销成功",
			Message: "软件已成功注销，请重新启动软件。",
		})
		//发送事件，通知前端更新注册状态
		rt.EventsEmit(app.ctx, "user:unregistered")
	})

	return appMenu
}
