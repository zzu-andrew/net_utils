package window

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/golang/glog"
	"github.com/zzu-andrew/net_utils/clipboard"
	"github.com/zzu-andrew/net_utils/resources"
	"log"
	"net/url"
	"strconv"
)

func shortcutFocused(s fyne.Shortcut, w fyne.Window) {
	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}

func (nu *NetUtils) NewMenu() *fyne.MainMenu {

	newItem := fyne.NewMenuItem("New", nil)
	otherItem := fyne.NewMenuItem("Other", nil)
	otherItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Project", func() { fmt.Println("Menu New->Other->Project") }),
		fyne.NewMenuItem("Mail", func() { fmt.Println("Menu New->Other->Mail") }),
	)
	newItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("File", func() { fmt.Println("Menu New->File") }),
		fyne.NewMenuItem("Directory", func() { fmt.Println("Menu New->Directory") }),
		otherItem,
	)
	settingsItem := fyne.NewMenuItem("Settings", func() {
		w := nu.GetApp().NewWindow("Fyne Settings")
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(480, 480))
		w.Show()
	})

	cutItem := fyne.NewMenuItem("Cut", func() {
		shortcutFocused(&fyne.ShortcutCut{
			Clipboard: nu.GetWin().Clipboard(),
		}, nu.GetWin())
	})
	copyItem := fyne.NewMenuItem("Copy", func() {
		shortcutFocused(&fyne.ShortcutCopy{
			Clipboard: nu.GetWin().Clipboard(),
		}, nu.GetWin())
	})
	pasteItem := fyne.NewMenuItem("Paste", func() {
		shortcutFocused(&fyne.ShortcutPaste{
			Clipboard: nu.GetWin().Clipboard(),
		}, nu.GetWin())
	})
	findItem := fyne.NewMenuItem("Find", func() { fmt.Println("Menu Find") })

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://developer.fyne.io")
			_ = nu.GetApp().OpenURL(u)
		}),
		fyne.NewMenuItem("Support", func() {
			u, _ := url.Parse("https://fyne.io/support/")
			_ = nu.GetApp().OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Sponsor", func() {
			u, _ := url.Parse("https://github.com/sponsors/fyne-io")
			_ = nu.GetApp().OpenURL(u)
		}))
	file := fyne.NewMenu("File", newItem)
	if !fyne.CurrentDevice().IsMobile() {
		file.Items = append(file.Items, fyne.NewMenuItemSeparator(), settingsItem)
	}
	return fyne.NewMainMenu(
		// a quit item will be appended to our first menu
		file,
		fyne.NewMenu("Edit", cutItem, copyItem, pasteItem, fyne.NewMenuItemSeparator(), findItem),
		helpMenu,
	)
}

//=======================================================================================================================================================================

type EtcdClient struct {
	ColorBoxAnimation *fyne.Animation
	CheckBoxAnimation *fyne.Animation
	InOutButton       *widget.Button
}

type KVConfig struct {
	k string
	v string
}

// Edit 不和应用绑定的窗口都放到Edit里面
type Edit struct {
	App             fyne.App
	Win             fyne.Window // 顶层窗口
	connEtcdDialog  dialog.Dialog
	leaseMngDialog  dialog.Dialog
	leaseSelect     *widget.Select
	clientIndex     int
	enableMirror    bool
	cli             [2]EtcdClient
	Tasks           *TaskApp
	Status          *widget.Label
	LeaseIDStatus   *widget.Label
	shortcutsDialog dialog.Dialog // 快捷键展示控件
	connectUsDialog dialog.Dialog
	doc             dialog.Dialog
	AddKV           KVConfig // 用于实时更新需要Add进ETcd的key value值
	leaseId         int64
}

func (ke *Edit) SetMirrorState(state bool) {
	ke.enableMirror = state
}

func (ke *Edit) GetMirrorState() bool {
	return ke.enableMirror
}

func (ke *Edit) GetInputButton() *widget.Button {
	return ke.cli[ke.clientIndex].InOutButton
}

func (ke *Edit) ConnEtcdForm() {
	if ke.connEtcdDialog == nil {
		// 首次进来将信息更改下
		ke.cli[0].InOutButton.SetText("Out")
		ke.cli[1].InOutButton.SetText("In")

		selectIndex := widget.NewSelect([]string{"0", "1"}, func(s string) {

		})
		selectIndex.SetSelected("0")

		username := widget.NewEntry()
		username.SetPlaceHolder("admin")
		//username.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]+$`, "username can only contain letters, numbers, '_', and '-'")
		password := widget.NewPasswordEntry()
		password.SetPlaceHolder("123456")
		//password.Validator = validation.NewRegexp(`^[A-Za-z0-9_-]+$`, "password can only contain letters, numbers, '_', and '-'")

		// 这里增加一个屏幕选择的窗口
		hostEntry := widget.NewEntry()
		//userNameEntry.Validator = validation.NewRegexp(`[1,2]`, "1 or 2 screen")
		// 设置预写字段
		//hostEntry.Validator = validation.NewRegexp(`/((ht|f)tps?:\/\/)?[\w-]+(\.[\w-]+)+:\d{1,5}\/?$/`,
		//	"Host Must contain ip port.")
		hostEntry.SetText("127.0.0.1:2379")
		// 设置占位符，虽然这里自己只有两个屏幕但是为了避免有很多屏幕的情况，还是选择使用10进制
		hostEntry.SetPlaceHolder("127.0.0.1:2379")
		remember := false
		items := []*widget.FormItem{
			widget.NewFormItem("Index", selectIndex),
			widget.NewFormItem("Username", username),
			widget.NewFormItem("Password", password),
			widget.NewFormItem("Host : ", hostEntry),
		}

		ke.connEtcdDialog = dialog.NewForm("Login...", "Log In", "Cancel", items, func(b bool) {
			if !b {
				return
			}
			var rememberText string
			if remember {
				rememberText = "and remember this login"
			}

			clientIndex, _ := strconv.Atoi(selectIndex.Selected)

			ke.cli[clientIndex].ColorBoxAnimation.Start()

			// 首次进来更新下信息，原先的信息是Start

			ke.cli[0].InOutButton.OnTapped = func() {
				if ke.cli[0].InOutButton.Text == "Out" {
					ke.cli[0].InOutButton.SetText("In")
					ke.cli[1].InOutButton.SetText("Out")
				} else {
					ke.cli[0].InOutButton.SetText("Out")
					ke.cli[1].InOutButton.SetText("In")
				}
			}
			ke.cli[1].InOutButton.OnTapped = func() {
				if ke.cli[1].InOutButton.Text == "In" {
					ke.cli[0].InOutButton.SetText("In")
					ke.cli[1].InOutButton.SetText("Out")
				} else {
					ke.cli[0].InOutButton.SetText("Out")
					ke.cli[1].InOutButton.SetText("In")
				}
			}

			log.Println("Please Authenticate", username.Text, password.Text, rememberText)
		}, ke.Win)
	}

	size := ke.Win.Canvas().Size()
	size.Width *= 0.70
	size.Height *= 0.70
	ke.connEtcdDialog.Resize(size)
	ke.connEtcdDialog.Show()
}

func (ke *Edit) ConfirmRadioData() {

	// 每次打开前先清空原先添加的radio
	ke.leaseSelect.Options = []string{"0"}
	// 将所有的lease添加到侯选项
	for _, lc := range ke.Tasks.LeaseVisible {
		if lc != nil {
			lease := strconv.FormatInt(lc.Lease, 10)
			ke.leaseSelect.Options = append(ke.leaseSelect.Options, lease)
		}
	}

}

func (ke *Edit) LeaseMngForm() {
	if ke.leaseMngDialog == nil {
		// 这里增加一个屏幕选择的窗口
		ke.leaseSelect = widget.NewSelect([]string{"0"}, func(s string) {

		})
		ke.leaseSelect.SetSelected("0")

		// 创建一个确认选择窗口
		ke.leaseMngDialog = dialog.NewCustomConfirm("Lease selection", "Confirm", "Dismiss",
			ke.leaseSelect, func(b bool) {
				if b {
					// 这里确保confirm之后再修改，如果没有confirm就不修改
					var err error
					ke.leaseId, err = strconv.ParseInt(ke.leaseSelect.Selected, 10, 64)
					if err != nil {
						ke.Status.SetText(err.Error())
						return
					}
					//	 如果确认成功，通过状态栏通知用户确认成功
					ke.LeaseIDStatus.SetText(ke.leaseSelect.Selected)
				}

			}, ke.Win)

	}

	// 显示窗口之前先将对应的内容更新下
	ke.ConfirmRadioData()
	size := ke.Win.Canvas().Size()
	size.Width *= 0.30
	size.Height *= 0.20
	ke.leaseMngDialog.Resize(size)
	ke.leaseMngDialog.Show()
}

// MakeNewMenu 创建菜单
// 1. 创建命令行菜单
// 2. 菜单相关控件初始化
// 3. 辅助信息初始化
func (ke *Edit) MakeNewMenu() *fyne.MainMenu {

	copyImage := fyne.NewMenuItem("CopyImage", func() {
		// 该窗口支持剪贴赋值
		ke.CopyImageToClip()
	})

	copyJson := fyne.NewMenuItem("CopyJson", func() {
		// 该窗口支持剪贴赋值
		if ke.Tasks.current != nil {
			marshal, err := json.Marshal(ke.Tasks.current)
			if err != nil {
				return
			}
			//ke.Win.Clipboard().SetContent(string(marshal))
			clipboard.CopyText(string(marshal))
		}
	})
	// 联系我们
	connectUsItem := fyne.NewMenuItem("Contact US", func() {
		ke.ConnectUsPage()
	})

	connectUsItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyH, Modifier: fyne.KeyModifierControl}
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Doc", func() {
			u, _ := url.Parse("https://gitee.com/andrewgithub/EtcdKeeperFyne/tree/master")
			_ = ke.App.OpenURL(u)
		}),
		fyne.NewMenuItem("Git Store", func() {
			u, _ := url.Parse("https://github.com/zzu-andrew/net_utils")
			//Open a URL in the default browser application.
			_ = ke.App.OpenURL(u)
		}),
		// 增加一个分割符号
		fyne.NewMenuItem("ShortCutInfo", func() {
			ke.showShortcutsPage()
		}),
		fyne.NewMenuItemSeparator(),
		connectUsItem,
	)

	host := fyne.NewMenuItem("Host", func() {
		ke.ConnEtcdForm()
	})

	mirror := fyne.NewMenuItem("Theme", func() {
		w := ke.App.NewWindow("Edit Settings")
		// creates a new settings screen to handle appearance configuration
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(480, 480))
		// Show the window on screen.
		w.Show()
	})

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Etcd", host, mirror),
		fyne.NewMenu("Edit", copyImage, fyne.NewMenuItemSeparator(), copyJson),
		helpMenu)
	//adds a top level menu to this window.
	return mainMenu
}

func (ke *Edit) ConnectUsPage() {

	// TODO: 支持展示多个网格形状图片
	if ke.connectUsDialog == nil {
		weChatImage := canvas.NewImageFromResource(resources.WeiChartIconPng)
		weChatContainer := container.NewScroll(weChatImage)
		weChatContainer.Resize(fyne.NewSize(420, 420))

		ke.connectUsDialog = dialog.NewCustom("EtcdKeeperFyne", "Confirm",
			weChatContainer,
			ke.Win)

	}

	size := ke.Win.Canvas().Size()
	size.Width *= 0.50
	size.Height *= size.Width
	ke.connectUsDialog.Resize(size)
	ke.connectUsDialog.Show()
}

// CopyImageToClip 将主窗口的画布转换成Image，并复制到剪贴板
func (ke *Edit) CopyImageToClip() {
	err := clipboard.CopyImage(ke.Win.Canvas().Capture())
	if err != nil {
		ke.Status.SetText(fmt.Sprintf("Copy image to clipboard failed. err : %s", err.Error()))
		return
	}
}

// RegisterShortcuts adds all the shortcuts and keys FireShotGO
// listens to.
// When updating here, please update also the `fs.ShowShortcutsPage()`
// method to reflect the changes.
func (ke *Edit) RegisterShortcuts() {
	// 注册Ctrl C 将复制的内容按照json字符创的形式拿出去
	ke.Win.Canvas().AddShortcut(
		&fyne.ShortcutCopy{},
		func(_ fyne.Shortcut) { ke.CopyImageToClip() })
	// 实际操作使用Control 调用窗口一类的使用Alt等
	ke.Win.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyL, Modifier: fyne.KeyModifierAlt},
		func(_ fyne.Shortcut) { ke.LeaseMngForm() })
	// Register shortcuts.
	ke.Win.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			glog.Infof("Quit requested by shortcut %s", shortcut.ShortcutName())
			ke.App.Quit()
		})

	ke.Win.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeySlash, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			glog.Infof("Update requested by shortcut %s", shortcut.ShortcutName())
			ke.showShortcutsPage()
		})

	ke.Win.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			glog.Infof("Marshal requested by shortcut %s", shortcut.ShortcutName())
			if ke.Tasks.current != nil {
				marshal, err := json.Marshal(ke.Tasks.current)
				if err != nil {
					return
				}
				ke.Win.Clipboard().SetContent(string(marshal))
			}
		})

	ke.Win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyEscape {
			ke.App.Quit()
		} else if ev.Name == fyne.KeyF5 {
			ke.Tasks.UpdateList()
			ke.Tasks.details.Refresh()
		} else {
			glog.V(2).Infof("KeyTyped: %+v", ev)
		}
	})
}

func (ke *Edit) showShortcutsPage() {
	if ke.shortcutsDialog == nil {
		titleFn := func(title string) (l *widget.Label) {
			l = widget.NewLabel(title)
			l.TextStyle.Bold = true
			return l
		}
		descFn := func(desc string) (l *widget.Label) {
			l = widget.NewLabel(desc)
			l.Alignment = fyne.TextAlignCenter
			return l
		}
		shortcutFn := func(shortcut string) (l *widget.Label) {
			l = widget.NewLabel(shortcut)
			l.TextStyle.Italic = true
			return l
		}
		ke.shortcutsDialog = dialog.NewCustom("EtcdKeeperFyne Shortcuts", "Ok",
			container.NewVScroll(container.NewVBox(
				titleFn("Data Manipulation"),
				container.NewGridWithColumns(2,
					descFn("Refresh Data"), shortcutFn("F5"),
					descFn("Select LeaseId"), shortcutFn("Alt+C"),
				),
				titleFn("Sharing Data"),
				container.NewGridWithColumns(2,
					descFn("Marshal Current Data"), shortcutFn("Control+M"),
					descFn("Copy Image To Clipboard"), shortcutFn("Control+C"),
				),
				titleFn("Other"),
				container.NewGridWithColumns(2,
					descFn("Shortcut page"), shortcutFn("Control+?"),
					descFn("Quit"), shortcutFn("Control+Q"),
				),
			)), ke.Win)
	}
	size := ke.Win.Canvas().Size()
	size.Width *= 0.80
	size.Height *= 0.80
	ke.shortcutsDialog.Resize(size)
	ke.shortcutsDialog.Show()
}
