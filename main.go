package main

import (
	"fmt"
	"github.com/zzu-andrew/net_utils/resources"
	"github.com/zzu-andrew/net_utils/window"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const preferenceCurrentTutorial = "currentTutorial"

var topWindow fyne.Window

func shortcutFocused(s fyne.Shortcut, w fyne.Window) {
	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}

func main() {

	netUtils := &window.NetUtils{
		App: app.NewWithID("net utils"),
	}

	netUtils.App.SetIcon(resources.ShotIconPng)
	w := netUtils.App.NewWindow("Net Utils")
	topWindow = w

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
		w := netUtils.App.NewWindow("Fyne Settings")
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(480, 480))
		w.Show()
	})

	cutItem := fyne.NewMenuItem("Cut", func() {
		shortcutFocused(&fyne.ShortcutCut{
			Clipboard: w.Clipboard(),
		}, w)
	})
	copyItem := fyne.NewMenuItem("Copy", func() {
		shortcutFocused(&fyne.ShortcutCopy{
			Clipboard: w.Clipboard(),
		}, w)
	})
	pasteItem := fyne.NewMenuItem("Paste", func() {
		shortcutFocused(&fyne.ShortcutPaste{
			Clipboard: w.Clipboard(),
		}, w)
	})
	findItem := fyne.NewMenuItem("Find", func() { fmt.Println("Menu Find") })

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://developer.fyne.io")
			_ = netUtils.App.OpenURL(u)
		}),
		fyne.NewMenuItem("Support", func() {
			u, _ := url.Parse("https://fyne.io/support/")
			_ = netUtils.App.OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Sponsor", func() {
			u, _ := url.Parse("https://github.com/sponsors/fyne-io")
			_ = netUtils.App.OpenURL(u)
		}))
	file := fyne.NewMenu("File", newItem)
	if !fyne.CurrentDevice().IsMobile() {
		file.Items = append(file.Items, fyne.NewMenuItemSeparator(), settingsItem)
	}
	mainMenu := fyne.NewMainMenu(
		// a quit item will be appended to our first menu
		file,
		fyne.NewMenu("Edit", cutItem, copyItem, pasteItem, fyne.NewMenuItemSeparator(), findItem),
		helpMenu,
	)
	w.SetMainMenu(mainMenu)
	w.SetMaster()

	content := container.NewMax()
	title := widget.NewLabel("Component name")
	intro := widget.NewLabel("An introduction would probably go\nhere, as well as a")
	intro.Wrapping = fyne.TextWrapWord
	setTutorial := func(t window.Tutorial) {

		title.SetText(t.Title)
		intro.SetText(t.Intro)

		content.Objects = []fyne.CanvasObject{t.View(netUtils, w)}
		content.Refresh()
	}

	tutorial := container.NewBorder(
		container.NewVBox(title, widget.NewSeparator(), intro), nil, nil, nil, content)

	split := container.NewHSplit(makeNav(setTutorial, true), tutorial)
	split.Offset = 0.2
	w.SetContent(split)

	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}

func makeNav(setTutorial func(tutorial window.Tutorial), loadPrevious bool) fyne.CanvasObject {
	a := fyne.CurrentApp()

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return window.TutorialIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := window.TutorialIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := window.Tutorials[uid]
			if !ok {
				fyne.LogError("Missing tutorial panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			if t, ok := window.Tutorials[uid]; ok {
				a.Preferences().SetString(preferenceCurrentTutorial, uid)
				setTutorial(t)
			}
		},
	}

	if loadPrevious {
		currentPref := a.Preferences().StringWithFallback(preferenceCurrentTutorial, "welcome")
		tree.Select(currentPref)
	}

	themes := container.New(layout.NewGridLayout(2),
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, tree)
}
