package tarpit

import (
	"bufio"
	"net"
	"net/http"
)

const (
	httpHeaderXForwardedFor = "X-Forwarded-For"
)

func getCallerIP(r *http.Request) string {
	ip := r.Header.Get(httpHeaderXForwardedFor)
	if ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func getURI(r *http.Request) string {
	return r.URL.RawPath
}

func hijack(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, nil, ErrHijackingUnsupported
	}
	conn, writeBuff, err := hj.Hijack()
	if err != nil {
		return nil, nil, err
	}
	return conn, writeBuff, nil
}
