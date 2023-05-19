package window

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"time"
)

type TaskApp struct {
	Ke       *Edit
	TaskData *TaskList
	// TODO: 记录当前代办事项，会存在多协程调用，需要加锁保护
	Visible []*task

	LeaseTaskData *TaskList
	LeaseVisible  []*task
	// 当前选中的task
	current      *task
	currentLease *task

	tasks     *widget.List
	details   *widget.Form
	leaseTask *widget.List
	//leaseShow   *widget.List
	leaseKVList *widget.List
	leaseKV     []string

	// Entry widget allows simple text to be input when focused.
	// 同一时间只展示一个元素的信息
	key            *widget.Entry
	value          *widget.Entry
	valueGrid      *widget.RichText
	due            *widget.Entry
	update         *widget.Button
	createRevision *widget.Label
	modRevision    *widget.Label
	version        *widget.Label
	lease          *widget.Label
}

func (a *TaskApp) refreshData() {
	// hide done
	a.Visible = a.TaskData.Remaining()
	a.tasks.Refresh()
}

func (a *TaskApp) refreshLeaseData() {
	// hide done
	a.LeaseVisible = a.LeaseTaskData.Remaining()
	a.leaseTask.Refresh()
}

// SetTask 更新当前界面上的数据信息
func (a *TaskApp) SetTask(t *task) {
	// 最新设置的那个task就是当前的代办列表
	// 当选中对应的list时会在这里指向对应的task
	a.current = t

	// 将data的title设置到APP的title中
	a.key.SetText(t.Key)
	a.value.SetText(t.Value)

	textSegMent := &widget.TextSegment{Style: widget.RichTextStyleInline, Text: t.Value}
	a.valueGrid.Segments = []widget.RichTextSegment{textSegMent}
	a.valueGrid.Refresh()

	a.createRevision.SetText(strconv.FormatInt(t.CreateRevision, 10))
	a.modRevision.SetText(strconv.FormatInt(t.ModRevision, 10))
	a.version.SetText(strconv.FormatInt(t.Version, 10))
	a.lease.SetText(strconv.FormatInt(t.Lease, 10))

	nowTime := time.Now()
	a.due.SetText(formatData(&nowTime))

	a.update.Refresh()
}

func formatData(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format(dateFormat)
}

//func makeTextGrid() *widget.TextGrid {
//	grid := widget.NewTextGrid()
//	grid.SetStyleRange(0, 4, 0, 7,
//		&widget.CustomTextGridStyle{BGColor: &color.NRGBA{R: 64, G: 64, B: 192, A: 128}})
//	grid.SetRowStyle(1, &widget.CustomTextGridStyle{BGColor: &color.NRGBA{R: 64, G: 192, B: 64, A: 128}})
//
//	grid.ShowLineNumbers = true
//	grid.ShowWhitespace = true
//
//	return grid
//}

func (a *TaskApp) NewKeyList() *widget.List {
	list := widget.NewList(
		func() int {
			return len(a.Visible)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewButtonWithIcon("Delete", theme.CancelIcon(), func() {}),
				widget.NewLabel("Etcd data"))
		},
		func(i widget.ListItemID, c fyne.CanvasObject) {
			task := a.Visible[i]
			box := c.(*fyne.Container)
			button := box.Objects[0].(*widget.Button)
			button.OnTapped = func() {
				// 这里将对应的key删除掉

				//a.details.Refresh()
				a.refreshData()
			}
			labelData := box.Objects[1].(*widget.Label)
			labelData.SetText(task.Key)
		})

	// 当选中一个list的时候，调用该回调函数
	// 选中哪个list之后，将界面上需要显示的都更换成当前ID 的list对象
	list.OnSelected = func(id widget.ListItemID) {
		a.SetTask(a.Visible[id])
	}
	return list
}

func (a *TaskApp) NewValueForm() *widget.Form {
	// 详细信息栏 key值，需要实时根据鼠标选择进行更新
	a.key = widget.NewEntry()
	// 详细信息栏 value值，需要实时根据鼠标选择进行更新
	a.value = widget.NewMultiLineEntry()
	// value值支持更新
	a.value.OnChanged = func(text string) {
		if a.current == nil {
			return
		}

		a.current.Value = text
	}

	a.createRevision = widget.NewLabel("0")
	a.modRevision = widget.NewLabel("0")
	a.version = widget.NewLabel("0")
	a.lease = widget.NewLabel("0")

	a.due = widget.NewEntry()
	a.due.Validator = dateValidator
	a.due.OnChanged = func(str string) {
		if a.current == nil {
			return
		}

		if str == "" {
			a.current.due = nil
		} else {
			date, err := time.Parse(dateFormat, str)
			if err != nil {
				a.current.due = &date
			}
		}
	}

	a.update = widget.NewButtonWithIcon("", theme.UploadIcon(), func() {
		//		当按下按钮的时候更新对应的value值

	})

	details := widget.NewForm(
		widget.NewFormItem("Key : ", a.key),
		widget.NewFormItem("Value : ", a.value),
		widget.NewFormItem("CreateRevision : ", a.createRevision),
		widget.NewFormItem("modRevision : ", a.modRevision),
		widget.NewFormItem("Version : ", a.version),
		widget.NewFormItem("Lease : ", a.lease),
		widget.NewFormItem("Due : ", a.due),
		widget.NewFormItem("UpdateValue : ", a.update),
	)
	return details
}

func (a *TaskApp) NewToolBar() *container.Split {
	//每次按更新这里都需要将最新的task进行更
	// 后期可以根据需要进行数据的实时更新，但是这样更加耗内存
	refresh := widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			a.UpdateList()
			a.details.Refresh()
		}),
	)

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("/xxx")
	keyEntry.OnChanged = func(text string) {
		a.Ke.AddKV.k = text
	}
	keyEntry.SetMinRowsVisible(20)

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("xxx")
	valueEntry.OnChanged = func(text string) {
		a.Ke.AddKV.v = text
	}
	valueEntry.SetMinRowsVisible(30)

	addConfig := container.New(&AddButtonLayout{}, widget.NewLabel("Key : "), keyEntry,
		widget.NewLabel("Value : "), valueEntry,
		widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
			// todo 将值的版本信息等都拿出来

			//a.TaskData.add(t)
			//a.SetTask(t)
			//// 刷新keyValue值
			//a.refreshData()
		}))

	ttlEntry := widget.NewEntry()
	ttlEntry.OnChanged = func(s string) {

	}
	ttlEntry.SetPlaceHolder("10")
	ttlEntry.SetText("10")
	grantLease := widget.NewToolbar(
		widget.NewToolbarAction(theme.MailAttachmentIcon(), func() {

			if len(ttlEntry.Text) == 0 || ttlEntry.Text == "0" {
				a.Ke.Status.SetText("ttl is 0")
				return
			}

			ttl, err := strconv.ParseInt(ttlEntry.Text, 10, 64)
			if err != nil {
				a.Ke.Status.SetText("convert ttl failed.")
				return
			}

			if err != nil {
				a.Ke.Status.SetText(fmt.Sprintf("Lease grant failed, err : %s", err.Error()))
				return
			}
			a.Ke.Status.SetText(fmt.Sprintf("Lease grant with ttl: %s", strconv.FormatInt(ttl, 10)))
		}),
	)

	toolBox := container.NewHBox(refresh, widget.NewLabel("TTL :"), ttlEntry, grantLease)

	// 为窗口添加中间的滑动窗口
	toolSplit := container.NewHSplit(
		toolBox,
		addConfig,
	)
	// 首个使用最小的空间
	toolSplit.Offset = 0.0
	return toolSplit
}

func (a *TaskApp) NewStatusBar() *fyne.Container {
	a.Ke.LeaseIDStatus = widget.NewLabel("0")
	return container.NewBorder(
		nil,
		nil,
		nil,
		container.NewHBox(widget.NewLabel("Lease ID: "), a.Ke.LeaseIDStatus),
		container.NewHBox(widget.NewLabel("Status : "), a.Ke.Status),
	)
}

func (a *TaskApp) NewLeaseList() *widget.List {
	return widget.NewList(
		func() int {
			return len(a.LeaseVisible)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewButtonWithIcon("KeepAlive", theme.HistoryIcon(), func() {}),
				widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {}),
				widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {}),
				widget.NewLabel("LeaseId"),
				widget.NewLabel("TTL"))
		},
		func(i widget.ListItemID, c fyne.CanvasObject) {
			lease := a.LeaseVisible[i]
			box := c.(*fyne.Container)
			keepAliveButton := box.Objects[0].(*widget.Button)
			refreshButton := box.Objects[1].(*widget.Button)
			deleteButton := box.Objects[2].(*widget.Button)
			leaseId := box.Objects[3].(*widget.Label)
			ttl := box.Objects[4].(*widget.Label)

			keepAliveButton.OnTapped = func() {

				// 不需要
				a.refreshLeaseData()
			}
			keepAliveButton.Importance = widget.HighImportance

			refreshButton.OnTapped = func() {
				// 对对应的LeaseId 进行包活

				// 不需要
				a.refreshLeaseData()
			}

			refreshButton.Importance = widget.HighImportance

			deleteButton.OnTapped = func() {
				// 这里将对应的key删除掉

				a.refreshLeaseData()
			}
			deleteButton.Importance = widget.DangerImportance
			// 首次构建窗口是使用
			leaseId.SetText(strconv.FormatInt(lease.Lease, 10))
			ttl.SetText(strconv.FormatInt(lease.TTL, 10))
		})
	// 当选中一个list的时候，调用该回调函数
	// 选中哪个list之后，将界面上需要显示的都更换成当前ID 的list对象
}

func (a *TaskApp) NewLeaseDetailShow() *container.Split {
	leaseShow := widget.NewList(
		func() int {
			return len(a.LeaseVisible)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("LeaseId"))
		},
		func(i widget.ListItemID, c fyne.CanvasObject) {
			lease := a.LeaseVisible[i]
			box := c.(*fyne.Container)
			leaseId := box.Objects[0].(*widget.Label)
			// 首次构建窗口是使用
			leaseId.SetText(strconv.FormatInt(lease.Lease, 10))
		})

	leaseShow.OnSelected = func(id widget.ListItemID) {
		//a.SetTask(a.LeaseVisible[id])
	}

	a.leaseKVList = widget.NewList(
		func() int {
			return len(a.leaseKV)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("lease keys"))
		},
		func(i widget.ListItemID, c fyne.CanvasObject) {
			kv := a.leaseKV[i]
			box := c.(*fyne.Container)
			kvLabel := box.Objects[0].(*widget.Label)
			kvLabel.SetText(kv)
		})
	return container.NewHSplit(leaseShow, a.leaseKVList)
}

func (a *TaskApp) NewMirrorSplit() *container.Split {
	//var aniScreen1 fyne.CanvasObject
	//aniScreen1, a.Ke.cli[0].ColorBoxAnimation, a.Ke.cli[0].CheckBoxAnimation, a.Ke.cli[0].InOutButton = MakeAnimationScreen()
	//
	//var aniScreen2 fyne.CanvasObject
	//aniScreen2, a.Ke.cli[1].ColorBoxAnimation, a.Ke.cli[1].CheckBoxAnimation, a.Ke.cli[1].InOutButton = MakeAnimationScreen()

	//makeMirrorButton := widget.NewButton("Make Mirror", func() {
	//
	//})
	//
	//radioHost := widget.NewRadioGroup([]string{"Host0", "Host1"}, func(s string) {
	//	switch s {
	//	case "Host0":
	//		a.Ke.clientIndex = 0
	//		a.UpdateList()
	//		a.details.Refresh()
	//	case "Host1":
	//		a.Ke.clientIndex = 1
	//		a.UpdateList()
	//		a.details.Refresh()
	//	}
	//
	//})
	//radioHost.SetSelected("Host0")
	//radioHost.Horizontal = true
	//makeMirror := container.NewHSplit(makeMirrorButton, radioHost)
	//
	//mirrorContainer := container.NewVSplit(makeMirror, container.NewHSplit(aniScreen1, aniScreen2))
	//mirrorContainer.Offset = 1
	return nil
}

func (a *TaskApp) MakeUI() fyne.CanvasObject {
	a.Ke.Status = widget.NewLabel("")
	a.current = &task{}
	// key值列表
	a.tasks = a.NewKeyList()
	// value详细信息
	a.details = a.NewValueForm()
	a.leaseTask = a.NewLeaseList()
	detailSplit := container.NewVSplit(container.NewVSplit(a.details, a.leaseTask),
		a.NewMirrorSplit())
	detailSplit.Offset = 1
	a.valueGrid = widget.NewRichTextWithText("")
	valueRichText := container.NewMax(container.NewVScroll(a.valueGrid))
	valueEdit := container.NewAppTabs(
		container.NewTabItem("Value Edit", detailSplit),
		container.NewTabItem("Value Show", valueRichText),
		container.NewTabItem("Lease", a.NewLeaseDetailShow()),
	)

	valueEdit.OnSelected = func(item *container.TabItem) {
		if valueEdit.SelectedIndex() == 1 {
			a.valueGrid.Refresh()
		}
	}
	// 为窗口添加中间的滑动窗口
	split := container.NewHSplit(
		a.tasks,
		valueEdit,
	)
	split.Offset = 0.3

	return container.NewBorder(a.NewToolBar(), a.NewStatusBar(), nil, nil, container.NewPadded(split))
}

func (a *TaskApp) UpdateTaskList() {

	a.refreshData()

}

func (a *TaskApp) UpdateList() {
	a.UpdateTaskList()
}

const (
	dateFormat = "02 Jan 06 15:04"

	lowPriority  = 0
	midPriority  = 1
	highPriority = 2
)

func dateValidator(text string) error {
	_, err := time.Parse(dateFormat, text)
	return err
}

type task struct {
	// create_revision is the revision of last creation on this key.
	CreateRevision int64 `json:"create_revision"`
	// mod_revision is the revision of last modification on this key.
	ModRevision int64 `json:"mod_revision"`
	// version is the version of the key. A deletion resets
	// the version to zero and any modification of the key
	// increases its version.
	Version int64 `json:"version"`
	// lease is the ID of the lease that attached to key.
	// When the attached lease expires, the key will be deleted.
	// If lease is 0, then no lease is attached to the key.
	Lease int64 `json:"lease"`
	// 对应leaseId的包活时间
	TTL        int64 `json:"ttl,omitempty"`
	GrantedTTL int64 `json:"granted_ttl,omitempty"`
	// 定义标题和描述语句
	Key   string `json:"key"`
	Value string `json:"value"`
	// 该任务是否已经完成
	done bool
	due  *time.Time
}

// TaskList 定义任务链表的切片
type TaskList struct {
	tasks []*task
}

// add 添加任务，新添加的任务放到列表头
func (l *TaskList) add(t *task) {
	// 使用t初始化一个一样的切片，并将原有的切片添加到后面
	l.tasks = append([]*task{t}, l.tasks...)
}

func (l *TaskList) del(t *task) {
	for i := 0; i < len(l.tasks); i++ {
		if l.tasks[i] == t {
			l.tasks = append(l.tasks[:i], l.tasks[i+1:]...)
			return
		}
	}
}

// Remaining 获取剩余没有完成的任务列表
func (l *TaskList) Remaining() []*task {
	var items []*task
	for _, task := range l.tasks {
		if !task.done {
			items = append(items, task)
		}
	}
	return items
}

/* 获取已经完成的任务列表 */
func (l *TaskList) done() []*task {
	var items []*task
	for _, task := range l.tasks {
		if task.done {
			items = append(items, task)
		}
	}
	return items
}

func EmptyData() *TaskList {
	return &TaskList{
		tasks: []*task{},
	}
}
