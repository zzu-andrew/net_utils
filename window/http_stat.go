package window

import (
	"encoding/json"
	"fmt"
	"github.com/zzu-andrew/net_utils/network"
	"net/url"

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

	//flag.Parse()
	httpStat := network.HttpStat("https://www.baidu.com")

	b, _ := json.Marshal(httpStat)
	fmt.Println(string(b))
	fmt.Println("==============================")

	uri := widget.NewLabel("")
	uri.SetText("")

	button := widget.NewButton("Conn", func() {

	})

	entryUri := widget.NewEntry()

	netUtils.httpStatObj = container.NewPadded(container.NewVSplit(
		container.NewHSplit(entryUri, button),
		widget.NewForm(
			widget.NewFormItem("Uri : ", uri),
			widget.NewFormItem("ConnectedTo : ", uri),
			widget.NewFormItem("ConnectedVia : ", uri),
			widget.NewFormItem("HttpInfo : ", uri),
			widget.NewFormItem("DnsLookup : ", uri),
			widget.NewFormItem("TcpConnection : ", uri),
			widget.NewFormItem("Due : ", uri),
			widget.NewFormItem("UpdateValue : ", uri),
		)))

	return netUtils.httpStatObj
}
