package window

import (
	"bytes"
	"context"
	"fmt"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/golang/glog"
	"github.com/zzu-andrew/net_utils/clipboard"
	"image"

	"github.com/wcharczuk/go-chart/v2"
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

		netUtils.httpStatInfo = network.HttpStat(ctx, uri, netUtils.status)

		uriWidget.SetText(netUtils.httpStatInfo.Uri)
		ConnectedToWidget.SetText(netUtils.httpStatInfo.Uri)
		ConnectedViaWidget.SetText(netUtils.httpStatInfo.ConnectedVia)
		HttpInfoWidget.SetText(netUtils.httpStatInfo.HttpInfo)
		//BodyWidget.SetText(httpStat.Body)
		DnsLookupWidget.SetText(netUtils.httpStatInfo.DnsLookup)
		TcpConnectionWidget.SetText(netUtils.httpStatInfo.TcpConnection)
		TlsHandshakeWidget.SetText(netUtils.httpStatInfo.TlsHandshake)
		ServerProcessingWidget.SetText(netUtils.httpStatInfo.ServerProcessing)
		ContentTransferWidget.SetText(netUtils.httpStatInfo.ContentTransfer)
		NameLookupWidget.SetText(netUtils.httpStatInfo.NameLookup)
		ConnectWidget.SetText(netUtils.httpStatInfo.Connect)
		PretransferWidget.SetText(netUtils.httpStatInfo.Pretransfer)
		StartTransferWidget.SetText(netUtils.httpStatInfo.StartTransfer)
		TotalWidget.SetText(netUtils.httpStatInfo.Total)
		BodyDiscardedWidget.SetText(netUtils.httpStatInfo.BodyDiscarded)

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

	t := widget.NewToolbar(
		widget.NewToolbarAction(theme.FileImageIcon(), func() {
			fmt.Println("New")
			showHttpStatImage(netUtils.win, &netUtils.httpStatInfo)
		}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			fmt.Println("Copy")
			copyHttpStatImageToClipboard(&netUtils.httpStatInfo)
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() { fmt.Println("Cut") }),

		widget.NewToolbarAction(theme.ContentPasteIcon(), func() { fmt.Println("Paste") }),
	)
	netUtils.httpStatObj = container.NewBorder(t, nil, nil, nil, httpStatPadd)

	return netUtils.httpStatObj
}

func showHttpStatImage(win fyne.Window, info *network.HttpStatInfo) {
	// TODO: 支持展示多个网格形状图片

	weChatImage := generatorHttpStatImage(info)
	weChatContainer := container.NewScroll(weChatImage)

	connectUsDialog := dialog.NewCustom("EtcdKeeperFyne", "Confirm",
		weChatContainer,
		win)

	size := win.Canvas().Size()
	size.Width = size.Width / 2
	size.Height = size.Height / 3
	connectUsDialog.Resize(size)
	connectUsDialog.Show()
}

func copyHttpStatImageToClipboard(info *network.HttpStatInfo) {

	buffer := generatorHttpStatPng(info)
	img, _, err := image.Decode(buffer)
	if err != nil {
		glog.Error(err.Error())
	}

	clipboard.CopyImage(img)
}

func generatorHttpStatImage(info *network.HttpStatInfo) *canvas.Image {
	return canvas.NewImageFromReader(generatorHttpStatPng(info), "")
}

const (
	httpsTemplate = `` +
		`  DNS Lookup   TCP Connection   TLS Handshake   Server Processing   Content Transfer` + "\n" +
		`[%s  |     %s  |    %s  |        %s  |       %s  ]` + "\n" +
		`            |                |               |                   |                  |` + "\n" +
		`   namelookup:%s      |               |                   |                  |` + "\n" +
		`                       connect:%s     |                   |                  |` + "\n" +
		`                                   pretransfer:%s         |                  |` + "\n" +
		`                                                     starttransfer:%s        |` + "\n" +
		`                                                                                total:%s` + "\n"

	httpTemplate = `` +
		`   DNS Lookup   TCP Connection   Server Processing   Content Transfer` + "\n" +
		`[ %s  |     %s  |        %s  |       %s  ]` + "\n" +
		`             |                |                   |                  |` + "\n" +
		`    namelookup:%s      |                   |                  |` + "\n" +
		`                        connect:%s         |                  |` + "\n" +
		`                                      starttransfer:%s        |` + "\n" +
		`                                                                 total:%s` + "\n"
)

func generatorHttpStatPng(info *network.HttpStatInfo) *bytes.Buffer {

	dnsTime, _ := strconv.Atoi(info.DnsLookup)
	tcpConnTime, _ := strconv.Atoi(info.TcpConnection)
	tlsTime, _ := strconv.Atoi(info.TlsHandshake)
	serTime, _ := strconv.Atoi(info.ServerProcessing)
	contentTransferTime, _ := strconv.Atoi(info.ContentTransfer)

	var yValues1 []float64
	if info.SchemeType == 1 {
		yValues1 = append(yValues1, float64(dnsTime))
		yValues1 = append(yValues1, float64(tcpConnTime))
		yValues1 = append(yValues1, float64(tlsTime))
		yValues1 = append(yValues1, float64(serTime))
		yValues1 = append(yValues1, float64(contentTransferTime))

	} else {

	}

	var xValues1 []float64
	for i := 0; i < len(yValues1); i++ {
		xValues1 = append(xValues1, float64(i+1))
	}

	ts1 := chart.ContinuousSeries{ //TimeSeries{
		Style: chart.Style{
			StrokeColor: chart.GetDefaultColor(2),
		},
		XValues: xValues1,
		YValues: yValues1,
	}

	ts2 := chart.ContinuousSeries{ //TimeSeries{
		Style: chart.Style{
			StrokeColor: chart.GetDefaultColor(1),
		},

		XValues: xValues1,
		YValues: yValues1,
	}

	graph := chart.Chart{

		XAxis: chart.XAxis{
			Name:           "The XAxis",
			ValueFormatter: chart.IntValueFormatter,
		},

		YAxis: chart.YAxis{
			Name:           "The YAxis",
			ValueFormatter: chart.IntValueFormatter,
		},

		Series: []chart.Series{
			ts1,
			ts2,
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
	}

	return buffer
}
