package window

import (
	"context"
	"fyne.io/fyne/v2/theme"
	"github.com/golang/glog"
	"github.com/zzu-andrew/net_utils/network"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

/*
 1. 每次从新获取uri
 2. 解析uri
 3. 将uri信息统计输出
*/

func httpStat(netUtils *NetUtils, _ fyne.Window) fyne.CanvasObject {

	if netUtils.httpStatObj != nil {
		return netUtils.httpStatObj
	}

	uriWidget := widget.NewLabel("")
	ConnectedToWidget := widget.NewLabel("")
	ConnectedViaWidget := widget.NewLabel("")
	HttpInfoWidget := widget.NewLabel("")
	//BodyWidget := widget.NewLabel("")
	DnsLookupWidget := widget.NewLabel("")
	TcpConnectionWidget := widget.NewLabel("")
	TlsHandshakeWidget := widget.NewLabel("")
	ServerProcessingWidget := widget.NewLabel("")
	ContentTransferWidget := widget.NewLabel("")
	NameLookupWidget := widget.NewLabel("")
	ConnectWidget := widget.NewLabel("")
	PretransferWidget := widget.NewLabel("")
	StartTransferWidget := widget.NewLabel("")
	TotalWidget := widget.NewLabel("")
	BodyDiscardedWidget := widget.NewLabel("")

	entryUri := widget.NewEntry()
	entryUri.SetPlaceHolder("https://www.baidu.com")

	timeOut := widget.NewEntry()

	timeOut.OnChanged = func(s string) {
		if _, err := strconv.Atoi(s); err != nil {
			glog.Error("Time out Entry string is invalid ", s)
			return
		}

		timeOut.Text = s
	}

	timeOut.Text = "10"
	addTimeButton := widget.NewButtonWithIcon("Add time(s)", theme.ContentAddIcon(), func() {
		st := timeOut.Text
		iTimeOut, err := strconv.Atoi(st)
		if err != nil {
			glog.Error("strconv.Atoi st failed.", err.Error())
		}

		iTimeOut += 1
		timeOut.SetText(strconv.Itoa(iTimeOut))
		timeOut.Refresh()
	})
	delTimeButton := widget.NewButtonWithIcon("Reduce time(s)", theme.ContentRemoveIcon(), func() {
		st := timeOut.Text
		iTimeOut, err := strconv.Atoi(st)
		if err != nil || iTimeOut <= 0 {
			glog.Error("strconv.Atoi st failed or timeOut is 0.", err.Error())
		}
		// 最低要求有1s的时间
		if iTimeOut <= 1 {
			return
		}

		iTimeOut -= 1
		timeOut.SetText(strconv.Itoa(iTimeOut))
		timeOut.Refresh()
	})

	button := widget.NewButton("Conn", func() {
		uri := entryUri.Text

		if len(uri) == 0 {
			return
		}
		var t int
		st := timeOut.Text
		if len(st) == 0 {
			glog.Info("timeOut is invalid set t = 10.")
			t = 10
		}
		t, err := strconv.Atoi(st)
		if err != nil {
			glog.Error("strconv.Atoi st : ", st, " failed.")
			t = 10
		}

		ctx, cancel := context.WithTimeout(netUtils.ctx, time.Second*time.Duration(t))
		defer cancel()

		httpStat := network.HttpStat(ctx, uri, netUtils.status)

		uriWidget.SetText(httpStat.Uri)
		ConnectedToWidget.SetText(httpStat.Uri)
		ConnectedViaWidget.SetText(httpStat.ConnectedVia)
		HttpInfoWidget.SetText(httpStat.HttpInfo)
		//BodyWidget.SetText(httpStat.Body)
		DnsLookupWidget.SetText(httpStat.DnsLookup)
		TcpConnectionWidget.SetText(httpStat.TcpConnection)
		TlsHandshakeWidget.SetText(httpStat.TlsHandshake)
		ServerProcessingWidget.SetText(httpStat.ServerProcessing)
		ContentTransferWidget.SetText(httpStat.ContentTransfer)
		NameLookupWidget.SetText(httpStat.NameLookup)
		ConnectWidget.SetText(httpStat.Connect)
		PretransferWidget.SetText(httpStat.Pretransfer)
		StartTransferWidget.SetText(httpStat.StartTransfer)
		TotalWidget.SetText(httpStat.Total)
		BodyDiscardedWidget.SetText(httpStat.BodyDiscarded)

	})

	timeBox := container.NewHSplit(container.NewHBox(timeOut, addTimeButton, delTimeButton), entryUri)
	timeBox.SetOffset(0.0)
	connectTool := container.NewHSplit(timeBox, button)
	// button 设置为最小
	connectTool.SetOffset(1.0)
	httpStatPadd := container.NewVSplit(
		connectTool,
		widget.NewForm(
			widget.NewFormItem("Uri : ", uriWidget),
			widget.NewFormItem("ConnectedTo : ", ConnectedToWidget),
			widget.NewFormItem("ConnectedVia : ", ConnectedViaWidget),
			widget.NewFormItem("HttpInfo : ", HttpInfoWidget),
			widget.NewFormItem("DnsLookup(ms) : ", DnsLookupWidget),
			widget.NewFormItem("TcpConnection(ms) : ", TcpConnectionWidget),
			widget.NewFormItem("TlsHandshake(ms) : ", TlsHandshakeWidget),
			widget.NewFormItem("ServerProcessing(ms) : ", ServerProcessingWidget),
			widget.NewFormItem("ContentTransfer(ms) : ", ContentTransferWidget),
			widget.NewFormItem("NameLookup(ms) : ", NameLookupWidget),
			widget.NewFormItem("Connect(ms) : ", ConnectWidget),
			widget.NewFormItem("Pretransfer(ms) : ", PretransferWidget),
			widget.NewFormItem("StartTransfer(ms) : ", StartTransferWidget),
			widget.NewFormItem("Total(ms) : ", TotalWidget),
			widget.NewFormItem("BodyDiscarded : ", BodyDiscardedWidget),
		))
	httpStatPadd.SetOffset(0.0)

	netUtils.httpStatObj = container.NewPadded(httpStatPadd)

	return netUtils.httpStatObj
}
