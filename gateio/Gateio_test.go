package gateio

import (
	"github.com/BTreeNewBee/goex"
	"github.com/BTreeNewBee/goex/internal/logger"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

var httpProxyClient = &http.Client{
	Transport: &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return &url.URL{
				Scheme: "socks5",
				Host:   "127.0.0.1:2341"}, nil
		},
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
	},
	Timeout: 10 * time.Second,
}

var (
	apikey    = ""
	secretkey = ""
)

//
var gateio *Gateio

func init() {
	logger.Log.SetLevel(logger.DEBUG)
	gateio = NewGateio(httpProxyClient, apikey, secretkey)
}

func TestGateio_GetAllCurrencyPair(t *testing.T) {
	t.Log(gateio.GetAllCurrencyPair())
}

func TestGateio_GetKLine(t *testing.T) {
	t.Log(gateio.GetKlineRecords(goex.BTC_USDT, goex.KLINE_PERIOD_1DAY, 1))
}

func TestGateio_GetAccount(t *testing.T) {
	//payload := "GET\n/api/v4/futures/orders\ncontract=BTC_USD&status=finished&limit=50\ncf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e\n1541993715"
	//t.Log(goex.GetParamHmacSHA512Sign("secret",payload))
	t.Log(gateio.GetAccount())
}
