package mxc

import (
	"fmt"
	"github.com/BTreeNewBee/goex"
	"github.com/BTreeNewBee/goex/internal/logger"
	"net"
	"net/http"
	"net/url"
	"strconv"
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
var mxc *Mxc

func init() {
	logger.Log.SetLevel(logger.DEBUG)
	mxc = NewMxc(httpProxyClient, apikey, secretkey)
}

func TestGateio_GetAllCurrencyPair(t *testing.T) {
	t.Log(mxc.GetAllCurrencyPair())
}

func TestGateio_GetTimestamp(t *testing.T) {
	t.Log(mxc.GetTimestamp())
}

func TestGateio_GetKLine(t *testing.T) {
	t.Log(mxc.GetKlineRecords(goex.BTC_USDT, goex.KLINE_PERIOD_1DAY, 2))
}

func TestMxc_GetUnfinishOrders(t *testing.T) {
	currency := goex.NewCurrencyPair(
		goex.NewCurrency("DOGGY", ""), goex.USDT,
	)
	t.Log(mxc.GetUnfinishOrders(currency))
}

func TestMxc_GetOneOrder(t *testing.T) {
	currency := goex.NewCurrencyPair(
		goex.NewCurrency("DOGGY", ""), goex.USDT,
	)
	t.Log(mxc.GetOneOrder("8cfd2b42aaf44d5db3e0108f8928031a", currency))
}

func TestMxc_MarketBuy(t *testing.T) {
	t.Log(mxc.GetAccount())
	currency := goex.NewCurrencyPair(
		goex.NewCurrency("DOGGY", ""), goex.USDT,
	)
	buy, err := mxc.MarketBuy("2000", "0.003", currency)
	t.Log(buy)
	t.Log(err)
	//t.Log(mxc.CancelOrder(buy.OrderID2, currency))
	t.Log(mxc.GetAccount())
}

func Test1(t *testing.T) {
	timestampStr := strconv.Itoa(int(time.Now().Unix()))
	fmt.Println(timestampStr)
	payload := fmt.Sprintf("%s&%s&%s", "mxcV9JCC8iu8zpaiWC", "1572936251", "api_key=mxcV9JCC8iu8zpaiWC&limit=1000&req_time=1572936251&startTime=1572076703000&symbol=MX_ETH&tradeType=BID")
	sign, _ := goex.GetParamHmacSHA256Sign(mxc.secretKey, payload)
	fmt.Println(sign)
}
