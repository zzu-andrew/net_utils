package window

import (
	"context"
	"github.com/zzu-andrew/net_utils/network"
	"net/url"
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

type HttpInfo struct {
	uri string // 需要进行测试的地址，支持http https
}

func parseURLa(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

/*

type HttpStatInfo struct {
	Uri              string            `json:"Url"`
	ConnectedTo      string            `json:"To"`
	ConnectedVia     string            `json:"Via"`
	HttpInfo         string            `json:"HttpInfo"`
	Body             map[string]string `json:"Body"`
	DnsLookup        string            `json:"DnsLookup"`
	TcpConnection    string            `json:"TcpConnection"`
	TlsHandshake     string            `json:"TlsHandshake"`
	ServerProcessing string            `json:"ServerProcessing"`
	ContentTransfer  string            `json:"ContentTransfer"`
	NameLookup       string            `json:"NameLookup"`
	Connect          string            `json:"Connect"`
	Pretransfer      string            `json:"Pretransfer"`
	StartTransfer    string            `json:"StartTransfer"`
	Total            string            `json:"Total"`
	BodyDiscarded    string            `json:"BodyDiscarded"`
}

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

	button := widget.NewButton("Conn", func() {
		uri := entryUri.Text

		if len(uri) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(netUtils.ctx, time.Second*1)
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
	connectTool := container.NewHSplit(entryUri, button)
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
