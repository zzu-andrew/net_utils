package network

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"fyne.io/fyne/v2/widget"
	"github.com/fatih/color"
	"github.com/golang/glog"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var (
	// Command line flags.
	httpMethod      string
	postBody        string
	followRedirects bool
	onlyHeader      bool
	insecure        bool
	httpHeaders     headers
	saveOutput      bool
	outputFile      string
	showVersion     bool
	clientCertFile  string
	fourOnly        bool
	sixOnly         bool

	// number of redirects followed
	redirectsFollowed int

	version = "devel" // for -v flag, updated during the release process with -ldflags=-X=main.version=...
)

const maxRedirects = 10

const (
	SchemeHttp  = 0
	SchemeHttps = 1
)

type HttpStatInfo struct {
	SchemeType       int               `json:"SchemeType"`
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

// HttpStat 获取http状态
func HttpStat(ctx context.Context, strUri string, status *widget.Label) HttpStatInfo {

	httpStatInfo := HttpStatInfo{
		Uri:  strUri,
		Body: make(map[string]string)}

	// url解析
	url := parseURL(strUri)

	httpStatInfo.visit(ctx, url, status)

	return httpStatInfo
}

// readClientCert - helper function to read client certificate
// from pem formatted file
func readClientCert(filename string) []tls.Certificate {
	if filename == "" {
		return nil
	}
	var (
		pkeyPem []byte
		certPem []byte
	)

	// read client certificate file (must include client private key and certificate)
	certFileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		glog.Error("failed to read client certificate file: %v", err)
	}

	for {
		block, rest := pem.Decode(certFileBytes)
		if block == nil {
			break
		}
		certFileBytes = rest

		if strings.HasSuffix(block.Type, "PRIVATE KEY") {
			pkeyPem = pem.EncodeToMemory(block)
		}
		if strings.HasSuffix(block.Type, "CERTIFICATE") {
			certPem = pem.EncodeToMemory(block)
		}
	}

	cert, err := tls.X509KeyPair(certPem, pkeyPem)
	if err != nil {
		glog.Error("unable to load client cert and key pair: %v", err)
	}
	return []tls.Certificate{cert}
}

func parseURL(uri string) *url.URL {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	url, err := url.Parse(uri)
	if err != nil {
		glog.Error("could not parse url %q: %v", uri, err)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
		if !strings.HasSuffix(url.Host, ":80") {
			url.Scheme += "s"
		}
	}
	return url
}

func headerKeyValue(h string) (string, string) {
	i := strings.Index(h, ":")
	if i == -1 {
		glog.Error("Header '%s' has invalid format, missing ':'", h)
	}
	return strings.TrimRight(h[:i], " "), strings.TrimLeft(h[i:], " :")
}

func dialContext(network string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, _, addr string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: false,
		}).DialContext(ctx, network, addr)
	}
}

// visit visits a url and times the interaction.
// If the response is a 30x, visit follows the redirect.
func (httpStatInfo *HttpStatInfo) visit(ctx context.Context, url *url.URL, status *widget.Label) {

	req := newRequest(httpMethod, url, postBody)
	// 这里定义计算使用的时间点
	var t0, t1, t2, t3, t4, t5, t6 time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1.IsZero() {
				// connecting to IP
				t1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				glog.Error("unable to connect to host %v: %v", addr, err)
			}
			t2 = time.Now()

			httpStatInfo.ConnectedTo = addr
		},
		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
		TLSHandshakeStart:    func() { t5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6 = time.Now() },
	}

	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	switch {
	case fourOnly:
		tr.DialContext = dialContext("tcp4")
	case sixOnly:
		tr.DialContext = dialContext("tcp6")
	}

	switch url.Scheme {
	case "https":
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}

		tr.TLSClientConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: insecure,
			Certificates:       readClientCert(clientCertFile),
			MinVersion:         tls.VersionTLS12,
		}
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// always refuse to follow redirects, visit does that
			// manually if required.
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		glog.Error("failed to read response: %v", err)
		status.SetText(fmt.Sprintf("failed to read response: %v", err))
		return
	}

	// Print SSL/TLS version which is used for connection
	connectedVia := "plaintext"
	if resp.TLS != nil {
		switch resp.TLS.Version {
		case tls.VersionTLS12:
			connectedVia = "TLSv1.2"
		case tls.VersionTLS13:
			connectedVia = "TLSv1.3"
		}
	}

	httpStatInfo.ConnectedVia = connectedVia

	resp.Body.Close()

	t7 := time.Now() // after read body
	if t0.IsZero() {
		// we skipped DNS
		t0 = t1
	}

	httpStatInfo.HttpInfo = fmt.Sprintf("%s%s%d.%d %s", "HTTP", "/", resp.ProtoMajor, resp.ProtoMinor, resp.Status)
	names := make([]string, 0, len(resp.Header))
	for k := range resp.Header {
		names = append(names, k)
	}
	sort.Sort(headers(names))
	for _, k := range names {
		httpStatInfo.Body[k] = strings.Join(resp.Header[k], ",")
	}

	fmtTime := func(d time.Duration) int {
		return int(d / time.Millisecond)
	}

	switch url.Scheme {
	case "https":
		httpStatInfo.SchemeType = 1
		httpStatInfo.DnsLookup = fmt.Sprintf("%d", fmtTime(t1.Sub(t0)))
		httpStatInfo.TcpConnection = fmt.Sprintf("%d", fmtTime(t2.Sub(t1)))    // tcp connection
		httpStatInfo.TlsHandshake = fmt.Sprintf("%d", fmtTime(t6.Sub(t5)))     // tls handshake
		httpStatInfo.ServerProcessing = fmt.Sprintf("%d", fmtTime(t4.Sub(t3))) // server processing
		httpStatInfo.ContentTransfer = fmt.Sprintf("%d", fmtTime(t7.Sub(t4)))  // content transfer
		httpStatInfo.NameLookup = fmt.Sprintf("%d", fmtTime(t1.Sub(t0)))       // namelookup
		httpStatInfo.Connect = fmt.Sprintf("%d", fmtTime(t2.Sub(t0)))          // connect
		httpStatInfo.Pretransfer = fmt.Sprintf("%d", fmtTime(t3.Sub(t0)))      // pretransfer
		httpStatInfo.StartTransfer = fmt.Sprintf("%d", fmtTime(t4.Sub(t0)))    // starttransfer
		httpStatInfo.Total = fmt.Sprintf("%d", fmtTime(t7.Sub(t0)))            // total

	case "http":
		httpStatInfo.SchemeType = 0
		httpStatInfo.DnsLookup = fmt.Sprintf("%d", fmtTime(t1.Sub(t0)))        // dns lookup
		httpStatInfo.TcpConnection = fmt.Sprintf("%d", fmtTime(t3.Sub(t1)))    // tcp connection
		httpStatInfo.ServerProcessing = fmt.Sprintf("%d", fmtTime(t4.Sub(t3))) // server processing
		httpStatInfo.ContentTransfer = fmt.Sprintf("%d", fmtTime(t7.Sub(t4)))  // content transfer
		httpStatInfo.NameLookup = fmt.Sprintf("%d", fmtTime(t1.Sub(t0)))       // namelookup
		httpStatInfo.Connect = fmt.Sprintf("%d", fmtTime(t3.Sub(t0)))          // connect
		httpStatInfo.Pretransfer = fmt.Sprintf("%d", fmtTime(t3.Sub(t0)))      // pretransfer
		httpStatInfo.StartTransfer = fmt.Sprintf("%d", fmtTime(t4.Sub(t0)))    // starttransfer
		httpStatInfo.Total = fmt.Sprintf("%d", fmtTime(t7.Sub(t0)))            // total

	}
	// 支持多重定向
	if followRedirects && isRedirect(resp) {
		loc, err := resp.Location()
		if err != nil {
			if err == http.ErrNoLocation {
				// 30x but no Location to follow, give up.
				return
			}
			glog.Error("unable to follow redirect: %v", err)
		}

		redirectsFollowed++
		if redirectsFollowed > maxRedirects {
			glog.Error("maximum number of redirects (%d) followed", maxRedirects)
		}

		httpStatInfo.visit(ctx, loc, status)
	}
}

func isRedirect(resp *http.Response) bool {
	return resp.StatusCode > 299 && resp.StatusCode < 400
}

func newRequest(method string, url *url.URL, body string) *http.Request {
	req, err := http.NewRequest(method, url.String(), createBody(body))
	if err != nil {
		glog.Error("unable to create request: %v", err)
	}
	for _, h := range httpHeaders {
		k, v := headerKeyValue(h)
		if strings.EqualFold(k, "host") {
			req.Host = v
			continue
		}
		req.Header.Add(k, v)
	}
	return req
}

func createBody(body string) io.Reader {
	if strings.HasPrefix(body, "@") {
		filename := body[1:]
		f, err := os.Open(filename)
		if err != nil {
			glog.Error("failed to open data file %s: %v", filename, err)
		}
		return f
	}
	return strings.NewReader(body)
}

// getFilenameFromHeaders tries to automatically determine the output filename,
// when saving to disk, based on the Content-Disposition header.
// If the header is not present, or it does not contain enough information to
// determine which filename to use, this function returns "".
func getFilenameFromHeaders(headers http.Header) string {
	// if the Content-Disposition header is set parse it
	if hdr := headers.Get("Content-Disposition"); hdr != "" {
		// pull the media type, and subsequent params, from
		// the body of the header field
		mt, params, err := mime.ParseMediaType(hdr)

		// if there was no error and the media type is attachment
		if err == nil && mt == "attachment" {
			if filename := params["filename"]; filename != "" {
				return filename
			}
		}
	}

	// return an empty string if we were unable to determine the filename
	return ""
}

// readResponseBody consumes the body of the response.
// readResponseBody returns an informational message about the
// disposition of the response body's contents.
func readResponseBody(req *http.Request, resp *http.Response) string {
	if isRedirect(resp) || req.Method == http.MethodHead {
		return ""
	}

	w := io.Discard
	msg := color.CyanString("Body discarded")

	if saveOutput || outputFile != "" {
		filename := outputFile

		if saveOutput {
			// try to get the filename from the Content-Disposition header
			// otherwise fall back to the RequestURI
			if filename = getFilenameFromHeaders(resp.Header); filename == "" {
				filename = path.Base(req.URL.RequestURI())
			}

			if filename == "/" {
				glog.Error("No remote filename; specify output filename with -o to save response body")
			}
		}

		f, err := os.Create(filename)
		if err != nil {
			glog.Error("unable to create file %s: %v", filename, err)
		}
		defer f.Close()
		w = f
		msg = color.CyanString("Body read")
	}

	if _, err := io.Copy(w, resp.Body); err != nil && w != ioutil.Discard {
		glog.Error("failed to read response body: %v", err)
	}

	return msg
}

type headers []string

func (h headers) String() string {
	var o []string
	for _, v := range h {
		o = append(o, "-H "+v)
	}
	return strings.Join(o, " ")
}

func (h *headers) Set(v string) error {
	*h = append(*h, v)
	return nil
}

func (h headers) Len() int      { return len(h) }
func (h headers) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h headers) Less(i, j int) bool {
	a, b := h[i], h[j]

	// server always sorts at the top
	if a == "Server" {
		return true
	}
	if b == "Server" {
		return false
	}

	endtoend := func(n string) bool {
		// https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.5.1
		switch n {
		case "Connection",
			"Keep-Alive",
			"Proxy-Authenticate",
			"Proxy-Authorization",
			"TE",
			"Trailers",
			"Transfer-Encoding",
			"Upgrade":
			return false
		default:
			return true
		}
	}

	x, y := endtoend(a), endtoend(b)
	if x == y {
		// both are of the same class
		return a < b
	}
	return x
}
