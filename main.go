package main

import (
	"flag"
	"fyne.io/fyne/v2/app"
	"github.com/golang/glog"
	"github.com/zzu-andrew/net_utils/resources"
	"github.com/zzu-andrew/net_utils/theme"
	window "github.com/zzu-andrew/net_utils/window"
)

// net utils工程，该工程用于执行linux命令，来监控linux 程序的性能

// 创建一个net utils的界面
func main() {
	// 参数解析
	flag.Parse()
	// 最后将日志进行更新
	defer glog.Flush()
	// 创建app对象
	edit := &window.Edit{
		App: app.NewWithID("net utils"),
	}
	// 设置程序图标
	edit.App.SetIcon(resources.ShotIconPng)
	// 更换主题，这样能支持中文？ 根据需要放开
	edit.App.Settings().SetTheme(&theme.Fzltch{RefThemeApp: edit.App,
		FontSizeName: "EtcdTheme"})
	// 创建主窗口，改窗户上完成所有工作
	edit.Win = edit.App.NewWindow("net utils")
	// 创建命令行菜单
	edit.Win.SetMainMenu(edit.MakeNewMenu())
	// 注册快捷键
	edit.RegisterShortcuts()

	data := window.EmptyData()
	edit.Tasks = &window.TaskApp{Ke: edit,
		TaskData:     data,
		Visible:      data.Remaining(),
		LeaseVisible: data.Remaining()}

	edit.Win.SetContent(edit.Tasks.MakeUI())

	edit.Win.ShowAndRun()
}
