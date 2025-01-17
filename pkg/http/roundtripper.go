package http

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"unicode/utf8"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type bearerRoundTripper struct {
	token   string
	wrapper http.RoundTripper
}

func NewBearerRoundTripper(token string, rt http.RoundTripper) http.RoundTripper {
	return &bearerRoundTripper{token: token, wrapper: rt}
}

func (rt *bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", rt.token))
	return rt.wrapper.RoundTrip(req)
}

type debugRoundTripper struct {
	next   http.RoundTripper
	logger log.Logger
}

func NewDebugRoundTripper(logger log.Logger, next http.RoundTripper) *debugRoundTripper {
	return &debugRoundTripper{next, log.With(logger, "component", "http/debugroundtripper")}
}

func (rt *debugRoundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	reqd, _ := httputil.DumpRequest(req, false)
	reqBody := bodyToString(&req.Body)

	res, err = rt.next.RoundTrip(req)
	if err != nil {
		level.Error(rt.logger).Log("err", err)
		return
	}

	resd, _ := httputil.DumpResponse(res, false)
	resBody := bodyToString(&res.Body)

	level.Debug(rt.logger).Log("msg", "round trip", "url", req.URL)
	level.Debug(rt.logger).Log("msg", "round trip", "requestdump", string(reqd))
	level.Debug(rt.logger).Log("msg", "round trip", "requestbody", reqBody)
	level.Debug(rt.logger).Log("msg", "round trip", "responsedump", string(resd))
	level.Debug(rt.logger).Log("msg", "round trip", "responsebody", resBody)
	return
}

func bodyToString(body *io.ReadCloser) string {
	if *body == nil {
		return "<nil>"
	}

	var b bytes.Buffer
	_, err := b.ReadFrom(*body)
	if err != nil {
		panic(err)
	}
	if err = (*body).Close(); err != nil {
		panic(err)
	}
	*body = ioutil.NopCloser(&b)

	s := b.String()
	if utf8.ValidString(s) {
		return s
	}

	return hex.Dump(b.Bytes())
}
