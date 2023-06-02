package window

import (
	"bytes"
	"context"
	"encoding/json"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/golang/glog"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/zzu-andrew/net_utils/clipboard"
	"github.com/zzu-andrew/net_utils/network"
	"image"
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

	updateForm := func() {
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
	}

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
		updateForm()
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
			widget.NewFormItem("Body : ", BodyDiscardedWidget),
		))
	httpStatPadd.SetOffset(0.0)

	t := widget.NewToolbar(
		widget.NewToolbarAction(theme.FileImageIcon(), func() {
			img := generatorHttpStatPng(&netUtils.httpStatInfo)
			if img == nil {
				return
			}
			showHttpStatImage(netUtils.win, img)
			clipboard.CopyImage(img)
		}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			copyHttpStatInfoJson(&netUtils.httpStatInfo)
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarSpacer(),
	)
	netUtils.httpStatObj = container.NewBorder(t, nil, nil, nil, httpStatPadd)

	return netUtils.httpStatObj
}

func showHttpStatImage(win fyne.Window, img image.Image) {

	connectUsDialog := dialog.NewCustom("HttpStat", "Confirm",
		container.NewScroll(canvas.NewImageFromImage(img)),
		win)

	size := win.Canvas().Size()
	size.Width = size.Width / 1.5
	size.Height = size.Height / 2.5
	connectUsDialog.Resize(size)
	connectUsDialog.Show()
}

func copyHttpStatInfoJson(info *network.HttpStatInfo) {

	buff, err := json.Marshal(info)
	if err != nil {
		glog.Error(err.Error())
		return
	}
	clipboard.CopyText(string(buff))
}

func generatorHttpStatPng(info *network.HttpStatInfo) image.Image {

	dnsTime, _ := strconv.Atoi(info.DnsLookup)
	tcpConnTime, _ := strconv.Atoi(info.TcpConnection)
	tlsTime, _ := strconv.Atoi(info.TlsHandshake)
	serTime, _ := strconv.Atoi(info.ServerProcessing)
	contentTransferTime, _ := strconv.Atoi(info.ContentTransfer)

	sumByIndex := func(xValue []float64, index int) float64 {
		var sum float64
		for i := 0; i < index; i++ {
			sum += xValue[i]
		}
		return sum
	}

	var xValues1, yValues2, yValues1 []float64
	var httpsValue2, httpsValue2Total []chart.Value2
	if info.SchemeType == network.SchemeHttps {

		var lables = []string{"DNS Lookup", "TCP Connection", "TLS Handshake", "Server Processing", "Content Transfer"}

		yValues1 = append(yValues1, float64(dnsTime))
		xValues1 = append(xValues1, float64(1))
		yValues1 = append(yValues1, float64(tcpConnTime))
		xValues1 = append(xValues1, float64(2))
		yValues1 = append(yValues1, float64(tlsTime))
		xValues1 = append(xValues1, float64(3))
		yValues1 = append(yValues1, float64(serTime))
		xValues1 = append(xValues1, float64(4))
		yValues1 = append(yValues1, float64(contentTransferTime))
		xValues1 = append(xValues1, float64(5))

		for i := 0; i < len(xValues1); i++ {
			httpsValue2 = append(httpsValue2, chart.Value2{XValue: xValues1[i], YValue: yValues1[i], Label: lables[i]})
		}

		for i := 0; i < len(xValues1); i++ {
			yValues2 = append(yValues2, sumByIndex(yValues1, i+1))
		}

		for i := 0; i < len(xValues1); i++ {
			httpsValue2Total = append(httpsValue2Total, chart.Value2{XValue: xValues1[i], YValue: yValues2[i], Label: lables[i]})
		}

	} else {
		var lables = []string{"DNS Lookup", "TCP Connection", "Server Processing", "Content Transfer"}

		yValues1 = append(yValues1, float64(dnsTime))
		xValues1 = append(xValues1, float64(1))
		yValues1 = append(yValues1, float64(tcpConnTime))
		xValues1 = append(xValues1, float64(2))
		yValues1 = append(yValues1, float64(serTime))
		xValues1 = append(xValues1, float64(3))
		yValues1 = append(yValues1, float64(contentTransferTime))
		xValues1 = append(xValues1, float64(4))

		for i := 0; i < len(xValues1); i++ {
			httpsValue2 = append(httpsValue2, chart.Value2{XValue: xValues1[i], YValue: yValues1[i], Label: lables[i]})
		}

		for i := 0; i < len(xValues1); i++ {
			yValues2 = append(yValues2, sumByIndex(yValues1, i+1))
		}

		for i := 0; i < len(xValues1); i++ {
			httpsValue2Total = append(httpsValue2Total, chart.Value2{XValue: xValues1[i], YValue: yValues2[i], Label: lables[i]})
		}
	}

	if 0 == sumByIndex(yValues1, len(yValues1)) {
		glog.Error("yValues1 is invalid, please conn first.")
		return nil
	}

	ts1 := chart.ContinuousSeries{ //TimeSeries{
		Name: "disperse",
		Style: chart.Style{
			StrokeColor: chart.GetDefaultColor(2),
		},
		XValues: xValues1,
		YValues: yValues1,
	}

	ts2 := chart.ContinuousSeries{ //TimeSeries{
		Name: "Total",
		Style: chart.Style{
			StrokeColor: chart.GetDefaultColor(1),
		},

		XValues: xValues1,
		YValues: yValues2,
	}

	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{
				Top:  20,
				Left: 20,
			},
		},
		XAxis: chart.XAxis{
			Name: "The Index",
		},

		YAxis: chart.YAxis{
			Name: "The length of time (ms)",
		},

		Series: []chart.Series{
			ts1,
			ts2,
			chart.AnnotationSeries{
				Annotations: httpsValue2,
			},
			chart.AnnotationSeries{
				Annotations: httpsValue2Total,
			},
		},
	}
	//note we have to do this as a separate step because we need a reference to graph
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
	}

	img, _, err := image.Decode(buffer)
	if err != nil {
		glog.Error(err.Error())
	}

	return img
}
