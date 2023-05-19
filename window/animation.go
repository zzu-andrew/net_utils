package window

import (
	"fyne.io/fyne/v2/container"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func MakeAnimationScreen() (fyne.CanvasObject, *fyne.Animation, *fyne.Animation, *widget.Button) {
	rect := canvas.NewRectangle(color.Black)
	rect.Resize(fyne.NewSize(320, 110))
	// 创建一个永久变动的色块，当成呼吸灯使用
	a := canvas.NewColorRGBAAnimation(theme.PrimaryColorNamed(theme.ColorBlue), theme.PrimaryColorNamed(theme.ColorGreen),
		time.Second*3, func(c color.Color) {
			rect.FillColor = c
			canvas.Refresh(rect)
		})
	a.RepeatCount = fyne.AnimationRepeatForever
	a.AutoReverse = true
	//检查按钮
	var a2 *fyne.Animation
	i := widget.NewIcon(theme.CheckButtonCheckedIcon())
	a2 = canvas.NewPositionAnimation(fyne.NewPos(0, 0), fyne.NewPos(260, 50), time.Second*3, func(p fyne.Position) {
		i.Move(p)

		width := 10 + (p.X / 7)
		i.Resize(fyne.NewSize(width, width))
	})
	a2.RepeatCount = fyne.AnimationRepeatForever
	a2.AutoReverse = true
	a2.Curve = fyne.AnimationLinear

	var toggle *widget.Button
	toggle = widget.NewButton("Start", func() {

	})
	toggle.Resize(toggle.MinSize())
	toggle.Move(fyne.NewPos(136, 37))
	return container.NewWithoutLayout(rect, i, toggle), a, a2, toggle
}
