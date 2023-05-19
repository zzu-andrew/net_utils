package window

import "fyne.io/fyne/v2"

// AddButtonLayout 用于添加按钮的布局设定
type AddButtonLayout struct {
}

func (c *AddButtonLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	incre := size.Width / float32(len(objs))
	count := 0
	if len(objs) == 5 {
		if size.Width > 700 {
			for _, child := range objs {
				// 偶数小一点，基数大一点
				if count%2 == 0 {
					if count == 4 {
						child.Resize(fyne.Size{Width: 80, Height: size.Height})
						child.Move(fyne.NewPos(x, 0))
						x += 70
					} else {
						child.Resize(fyne.Size{Width: 50, Height: size.Height})
						child.Move(fyne.NewPos(x, 0))
						x += 50
					}

				} else {
					ss := (size.Width - 150) / 2
					if ss > 260 {
						ss = 260
					}
					child.Resize(fyne.Size{Width: ss, Height: size.Height})
					child.Move(fyne.NewPos(x, 0))
					x += ss
				}
				count++
			}
		} else {
			for _, child := range objs {
				child.Resize(fyne.Size{Width: incre, Height: size.Height})
				child.Move(fyne.NewPos(x, 0))
				x += incre
			}
		}

	} else {
		for _, child := range objs {
			child.Resize(fyne.Size{Width: incre, Height: size.Height})
			child.Move(fyne.NewPos(x, 0))
			x += incre
		}
	}

}

func (c *AddButtonLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(300, 36)
}
