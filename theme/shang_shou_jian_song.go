package theme

import (
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

// Fzltch 设置自定义主题，主要用于支持中文
type Fzltch struct {
	RefThemeApp  fyne.App
	FontSizeName string
}

// ShangShouJianSongXianXiTi 1. 第一种方式
// 这个功能只有go 1.16之后的版本才支持的，如果你的版本是1.16之前，请使用
// fyne bundle fzltzch-2.ttf > bundle.go
// 2. 第二种方式
//
//go:embed fzltzch.TTF
var ShangShouJianSongXianXiTi []byte

var resourceShangShouJianSongXianXiTi2Ttf = &fyne.StaticResource{
	StaticName:    "ShangShouJianSongXianXiTi-2.ttf",
	StaticContent: ShangShouJianSongXianXiTi,
}

// Font 返回的就是字体名
func (sm *Fzltch) Font(s fyne.TextStyle) fyne.Resource {

	if s.Monospace || s.Bold || s.Italic {
		return theme.DefaultTheme().Font(s)
	}

	return resourceShangShouJianSongXianXiTi2Ttf
}

func (*Fzltch) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}

func (*Fzltch) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (sm *Fzltch) Size(n fyne.ThemeSizeName) float32 {

	return theme.DefaultTheme().Size(n)
}
