package resources

// This file embeds all the resources used by the program.

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed fire.png
var embedIconPng []byte
var ShotIconPng = fyne.NewStaticResource("ShotIconPng", embedIconPng)

//go:embed weixin.png
var weixinIconPng []byte //
var WeiChartIconPng = fyne.NewStaticResource("WeiChartIconPng", weixinIconPng)
