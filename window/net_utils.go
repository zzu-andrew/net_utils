package window

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/zzu-andrew/net_utils/resources"
)

type UtilsData struct {
}

type NetUtils struct {
	ctx         context.Context
	app         fyne.App
	win         fyne.Window
	status      *widget.Label
	broadcast   *widget.Label
	httpStatObj fyne.CanvasObject
}

func NewNetUtils() *NetUtils {
	nu := &NetUtils{
		ctx: context.Background(),
		app: app.NewWithID("net utils"),
	}

	nu.GetApp().SetIcon(resources.ShotIconPng)
	nu.SetWin(nu.GetApp().NewWindow("Net Utils"))

	nu.GetWin().SetMainMenu(nu.NewMenu())
	nu.GetWin().SetMaster()
	return nu
}

// SetApp 外部接口使用
func (nu *NetUtils) SetApp(app fyne.App) {
	nu.app = app
}

func (nu *NetUtils) GetApp() fyne.App {
	return nu.app
}

func (nu *NetUtils) GetWin() fyne.Window {
	return nu.win
}

func (nu *NetUtils) SetWin(win fyne.Window) {
	nu.win = win
}
