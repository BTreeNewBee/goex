package gateio

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/BTreeNewBee/goex"
	. "github.com/BTreeNewBee/goex/internal/logger"
)

var _INERNAL_KLINE_PERIOD_CONVERTER = map[KlinePeriod]string{
	KLINE_PERIOD_1MIN:   "1m",
	KLINE_PERIOD_3MIN:   "3m",
	KLINE_PERIOD_5MIN:   "5m",
	KLINE_PERIOD_15MIN:  "15m",
	KLINE_PERIOD_30MIN:  "30m",
	KLINE_PERIOD_60MIN:  "1h",
	KLINE_PERIOD_1H:     "1h",
	KLINE_PERIOD_2H:     "2h",
	KLINE_PERIOD_4H:     "4h",
	KLINE_PERIOD_6H:     "6h",
	KLINE_PERIOD_8H:     "8h",
	KLINE_PERIOD_12H:    "12h",
	KLINE_PERIOD_1DAY:   "1d",
	KLINE_PERIOD_3DAY:   "3d",
	KLINE_PERIOD_1WEEK:  "7d",
	KLINE_PERIOD_1MONTH: "1M",
}

type AccountInfo struct {
	Id    string
	Type  string
	State string
}

type Gateio struct {
	httpClient *http.Client
	baseUrl    string
	accountId  string
	accessKey  string
	secretKey  string
}

func NewGateioWithConfig(config *APIConfig) *Gateio {
	gateio := new(Gateio)
	if config.Endpoint == "" {
		gateio.baseUrl = "https://api.gateio.ws"
	} else {
		gateio.baseUrl = config.Endpoint
	}
	gateio.httpClient = config.HttpClient
	gateio.accessKey = config.ApiKey
	gateio.secretKey = config.ApiSecretKey
	return gateio
}

func NewGateio(httpClient *http.Client, apiKey string, apiSecretKey string) *Gateio {
	gateio := new(Gateio)
	gateio.baseUrl = "https://api.gateio.ws"
	gateio.httpClient = httpClient
	gateio.accessKey = apiKey
	gateio.secretKey = apiSecretKey
	return gateio
}

func (gateio *Gateio) GetAccountInfo(acc string) (AccountInfo, error) {
	path := "/wallet/sub_account_balances"
	params := &url.Values{}
	headers := map[string]string{}
	gateio.buildSign("GET", path, "", &headers)

	log.Println(gateio.baseUrl + path + "?" + params.Encode())
	//log.Println(gateio.baseUrl + path + "?" + params.Encode())

	respmap, err := HttpGet3(gateio.httpClient, gateio.baseUrl+path+"?"+params.Encode(), headers)
	if err != nil {
		return AccountInfo{}, err
	}

	var info AccountInfo

	for _, v := range respmap {
		iddata := v.(map[string]interface{})
		if iddata["type"].(string) == acc {
			info.Id = fmt.Sprintf("%.0f", iddata["id"])
			info.Type = acc
			info.State = iddata["state"].(string)
			break
		}
	}
	//log.Println(respmap)
	return info, nil
}

func (gateio *Gateio) GetAccount() (*Account, error) {
	path := "/api/v4/spot/accounts"
	headers := map[string]string{}
	gateio.buildSign("GET", path, "", &headers)

	//log.Println(gateio.baseUrl + path + "?" + params.Encode())

	respmap, err := HttpGet3(gateio.httpClient, gateio.baseUrl+path, headers)
	if err != nil {
		return nil, err
	}

	acc := new(Account)
	acc.SubAccounts = make(map[Currency]SubAccount, 6)
	acc.Exchange = gateio.GetExchangeName()

	for _, v := range respmap {
		v1 := v.(map[string]interface{})
		s := new(SubAccount)
		s.ForzenAmount = ToFloat64(v1["locked"])
		s.LoanAmount = 0
		s.Amount = ToFloat64(v1["available"])
		s.Currency = NewCurrency(v1["currency"].(string), "")
		acc.SubAccounts[s.Currency] = *s
	}

	return acc, nil
}

func (gateio *Gateio) placeOrder(amount, price string, pair CurrencyPair, orderType string) (string, error) {
	//symbol := gateio.Symbols[pair.ToLower().ToSymbol("")]
	//
	//path := "/v1/order/orders/place"
	//params := url.Values{}
	//params.Set("account-id", gateio.accountId)
	//params.Set("client-order-id", GenerateOrderClientId(32))
	//params.Set("amount", FloatToString(ToFloat64(amount), int(symbol.AmountPrecision)))
	//params.Set("symbol", pair.AdaptUsdToUsdt().ToLower().ToSymbol(""))
	//params.Set("type", orderType)
	//
	//switch orderType {
	//case "buy-limit", "sell-limit":
	//	params.Set("price", FloatToString(ToFloat64(price), int(symbol.PricePrecision)))
	//}
	//
	//gateio.buildPostForm("POST", path, &params)
	//
	//resp, err := HttpPostForm3(gateio.httpClient, gateio.baseUrl+path+"?"+params.Encode(), gateio.toJson(params),
	//	map[string]string{"Content-Type": "application/json", "Accept-Language": "zh-cn"})
	//if err != nil {
	//	return "", err
	//}
	//
	//respmap := make(map[string]interface{})
	//err = json.Unmarshal(resp, &respmap)
	//if err != nil {
	//	return "", err
	//}
	//
	//if respmap["status"].(string) != "ok" {
	//	return "", errors.New(respmap["err-code"].(string))
	//}

	return "", nil
}

func (gateio *Gateio) LimitBuy(amount, price string, currency CurrencyPair, opt ...LimitOrderOptionalParameter) (*Order, error) {
	orderTy := "buy-limit"
	if len(opt) > 0 {
		switch opt[0] {
		case PostOnly:
			orderTy = "buy-limit-maker"
		case Ioc:
			orderTy = "buy-ioc"
		case Fok:
			orderTy = "buy-limit-fok"
		default:
			Log.Error("limit order optional parameter error ,opt= ", opt[0])
		}
	}
	orderId, err := gateio.placeOrder(amount, price, currency, orderTy)
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     BUY}, nil
}

func (gateio *Gateio) LimitSell(amount, price string, currency CurrencyPair, opt ...LimitOrderOptionalParameter) (*Order, error) {
	orderTy := "sell-limit"
	if len(opt) > 0 {
		switch opt[0] {
		case PostOnly:
			orderTy = "sell-limit-maker"
		case Ioc:
			orderTy = "sell-ioc"
		case Fok:
			orderTy = "sell-limit-fok"
		default:
			Log.Error("limit order optional parameter error ,opt= ", opt[0])
		}
	}
	orderId, err := gateio.placeOrder(amount, price, currency, orderTy)
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     SELL}, nil
}

func (gateio *Gateio) MarketBuy(amount, price string, currency CurrencyPair) (*Order, error) {
	orderId, err := gateio.placeOrder(amount, price, currency, "buy-market")
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     BUY_MARKET}, nil
}

func (gateio *Gateio) MarketSell(amount, price string, currency CurrencyPair) (*Order, error) {
	orderId, err := gateio.placeOrder(amount, price, currency, "sell-market")
	if err != nil {
		return nil, err
	}
	return &Order{
		Currency: currency,
		OrderID:  ToInt(orderId),
		OrderID2: orderId,
		Amount:   ToFloat64(amount),
		Price:    ToFloat64(price),
		Side:     SELL_MARKET}, nil
}

func (gateio *Gateio) parseOrder(ordmap map[string]interface{}) Order {
	ord := Order{
		Cid:        fmt.Sprint(ordmap["client-order-id"]),
		OrderID:    ToInt(ordmap["id"]),
		OrderID2:   fmt.Sprint(ToInt(ordmap["id"])),
		Amount:     ToFloat64(ordmap["amount"]),
		Price:      ToFloat64(ordmap["price"]),
		DealAmount: ToFloat64(ordmap["field-amount"]),
		Fee:        ToFloat64(ordmap["field-fees"]),
		OrderTime:  ToInt(ordmap["created-at"]),
	}

	state := ordmap["state"].(string)
	switch state {
	case "submitted", "pre-submitted":
		ord.Status = ORDER_UNFINISH
	case "filled":
		ord.Status = ORDER_FINISH
	case "partial-filled":
		ord.Status = ORDER_PART_FINISH
	case "canceled", "partial-canceled":
		ord.Status = ORDER_CANCEL
	default:
		ord.Status = ORDER_UNFINISH
	}

	if ord.DealAmount > 0.0 {
		ord.AvgPrice = ToFloat64(ordmap["field-cash-amount"]) / ord.DealAmount
	}

	typeS := ordmap["type"].(string)
	switch typeS {
	case "buy-limit":
		ord.Side = BUY
	case "buy-market":
		ord.Side = BUY_MARKET
	case "sell-limit":
		ord.Side = SELL
	case "sell-market":
		ord.Side = SELL_MARKET
	}
	return ord
}

func (gateio *Gateio) GetOneOrder(orderId string, currency CurrencyPair) (*Order, error) {
	path := "/v1/order/orders/" + orderId
	params := url.Values{}
	gateio.buildPostForm("GET", path, &params)
	respmap, err := HttpGet(gateio.httpClient, gateio.baseUrl+path+"?"+params.Encode())
	if err != nil {
		return nil, err
	}

	if respmap["status"].(string) != "ok" {
		return nil, errors.New(respmap["err-code"].(string))
	}

	datamap := respmap["data"].(map[string]interface{})
	order := gateio.parseOrder(datamap)
	order.Currency = currency

	return &order, nil
}

func (gateio *Gateio) GetUnfinishOrders(currency CurrencyPair) ([]Order, error) {
	return gateio.getOrders(currency, OptionalParameter{}.
		Optional("states", "pre-submitted,submitted,partial-filled").
		Optional("size", "100"))
}

func (gateio *Gateio) CancelOrder(orderId string, currency CurrencyPair) (bool, error) {
	path := fmt.Sprintf("/v1/order/orders/%s/submitcancel", orderId)
	params := url.Values{}
	gateio.buildPostForm("POST", path, &params)
	resp, err := HttpPostForm3(gateio.httpClient, gateio.baseUrl+path+"?"+params.Encode(), gateio.toJson(params),
		map[string]string{"Content-Type": "application/json", "Accept-Language": "zh-cn"})
	if err != nil {
		return false, err
	}

	var respmap map[string]interface{}
	err = json.Unmarshal(resp, &respmap)
	if err != nil {
		return false, err
	}

	if respmap["status"].(string) != "ok" {
		return false, errors.New(string(resp))
	}

	return true, nil
}

func (gateio *Gateio) GetOrderHistorys(currency CurrencyPair, optional ...OptionalParameter) ([]Order, error) {
	var optionals []OptionalParameter
	optionals = append(optionals, OptionalParameter{}.
		Optional("states", "canceled,partial-canceled,filled").
		Optional("size", "100").
		Optional("direct", "next"))
	optionals = append(optionals, optional...)
	return gateio.getOrders(currency, optionals...)
}

type queryOrdersParams struct {
	types,
	startDate,
	endDate,
	states,
	from,
	direct string
	size int
	pair CurrencyPair
}

func (gateio *Gateio) getOrders(pair CurrencyPair, optional ...OptionalParameter) ([]Order, error) {
	path := "/v1/order/orders"
	params := url.Values{}
	params.Set("symbol", strings.ToLower(pair.AdaptUsdToUsdt().ToSymbol("")))
	MergeOptionalParameter(&params, optional...)
	Log.Info(params)
	gateio.buildPostForm("GET", path, &params)
	respmap, err := HttpGet(gateio.httpClient, fmt.Sprintf("%s%s?%s", gateio.baseUrl, path, params.Encode()))
	if err != nil {
		return nil, err
	}

	if respmap["status"].(string) != "ok" {
		return nil, errors.New(respmap["err-code"].(string))
	}

	datamap := respmap["data"].([]interface{})
	var orders []Order
	for _, v := range datamap {
		ordmap := v.(map[string]interface{})
		ord := gateio.parseOrder(ordmap)
		ord.Currency = pair
		orders = append(orders, ord)
	}

	return orders, nil
}

func (gateio *Gateio) GetTicker(currencyPair CurrencyPair) (*Ticker, error) {
	pair := currencyPair.AdaptUsdToUsdt()
	url := gateio.baseUrl + "/market/detail/merged?symbol=" + strings.ToLower(pair.ToSymbol(""))
	respmap, err := HttpGet(gateio.httpClient, url)
	if err != nil {
		return nil, err
	}

	if respmap["status"].(string) == "error" {
		return nil, errors.New(respmap["err-msg"].(string))
	}

	tickmap, ok := respmap["tick"].(map[string]interface{})
	if !ok {
		return nil, errors.New("tick assert error")
	}

	ticker := new(Ticker)
	ticker.Pair = currencyPair
	ticker.Vol = ToFloat64(tickmap["amount"])
	ticker.Low = ToFloat64(tickmap["low"])
	ticker.High = ToFloat64(tickmap["high"])
	bid, isOk := tickmap["bid"].([]interface{})
	if isOk != true {
		return nil, errors.New("no bid")
	}
	ask, isOk := tickmap["ask"].([]interface{})
	if isOk != true {
		return nil, errors.New("no ask")
	}
	ticker.Buy = ToFloat64(bid[0])
	ticker.Sell = ToFloat64(ask[0])
	ticker.Last = ToFloat64(tickmap["close"])
	ticker.Date = ToUint64(respmap["ts"])

	return ticker, nil
}

func (gateio *Gateio) GetDepth(size int, currency CurrencyPair) (*Depth, error) {
	url := gateio.baseUrl + "/market/depth?symbol=%s&type=step0&depth=%d"
	n := 5
	pair := currency.AdaptUsdToUsdt()
	if size <= 5 {
		n = 5
	} else if size <= 10 {
		n = 10
	} else if size <= 20 {
		n = 20
	} else {
		url = gateio.baseUrl + "/market/depth?symbol=%s&type=step0&d=%d"
	}
	respmap, err := HttpGet(gateio.httpClient, fmt.Sprintf(url, strings.ToLower(pair.ToSymbol("")), n))
	if err != nil {
		return nil, err
	}

	if "ok" != respmap["status"].(string) {
		return nil, errors.New(respmap["err-msg"].(string))
	}

	tick, _ := respmap["tick"].(map[string]interface{})

	dep := gateio.parseDepthData(tick, size)
	dep.Pair = currency
	mills := ToUint64(tick["ts"])
	dep.UTime = time.Unix(int64(mills/1000), int64(mills%1000)*int64(time.Millisecond))

	return dep, nil
}

//倒序
func (gateio *Gateio) GetKlineRecords(currency CurrencyPair, period KlinePeriod, size int, optional ...OptionalParameter) ([]Kline, error) {
	url := gateio.baseUrl + "/api/v4/spot/candlesticks?interval=%s&limit=%d&currency_pair=%s"
	symbol := strings.ToLower(currency.AdaptUsdToUsdt().ToSymbol("_"))
	periodS, isOk := _INERNAL_KLINE_PERIOD_CONVERTER[period]
	if isOk != true {
		periodS = "1d"
	}

	ret, err := HttpGet3(gateio.httpClient, fmt.Sprintf(url, periodS, size, symbol), map[string]string{})
	if err != nil {
		return nil, err
	}

	var klines []Kline
	for _, e := range ret {
		item := e.([]interface{})
		klines = append(klines, Kline{
			Pair:      currency,
			Open:      ToFloat64(item[5]),
			Close:     ToFloat64(item[2]),
			High:      ToFloat64(item[3]),
			Low:       ToFloat64(item[4]),
			Vol:       ToFloat64(item[1]),
			Timestamp: int64(ToUint64(item[0]))})
	}

	return klines, nil
}

func (gateio *Gateio) GetTrades(currencyPair CurrencyPair, since int64) ([]Trade, error) {
	var (
		trades []Trade
		ret    struct {
			Status string
			ErrMsg string `json:"err-msg"`
			Data   []struct {
				Ts   int64
				Data []struct {
					Id        big.Int
					Amount    float64
					Price     float64
					Direction string
					Ts        int64
				}
			}
		}
	)

	url := gateio.baseUrl + "/market/history/trade?size=2000&symbol=" + currencyPair.AdaptUsdToUsdt().ToLower().ToSymbol("")
	err := HttpGet4(gateio.httpClient, url, map[string]string{}, &ret)
	if err != nil {
		return nil, err
	}

	if ret.Status != "ok" {
		return nil, errors.New(ret.ErrMsg)
	}

	for _, d := range ret.Data {
		for _, t := range d.Data {

			//fix huobi   Weird rules of tid
			//火币交易ID规定固定23位, 导致超出int64范围，每个交易对有不同的固定填充前缀
			//实际交易ID远远没有到23位数字。
			tid := ToInt64(strings.TrimPrefix(t.Id.String()[4:], "0"))
			if tid == 0 {
				tid = ToInt64(strings.TrimPrefix(t.Id.String()[5:], "0"))
			}
			///

			trades = append(trades, Trade{
				Tid:    ToInt64(tid),
				Pair:   currencyPair,
				Amount: t.Amount,
				Price:  t.Price,
				Type:   AdaptTradeSide(t.Direction),
				Date:   t.Ts})
		}
	}

	return trades, nil
}

type ecdsaSignature struct {
	R, S *big.Int
}

func (gateio *Gateio) buildSign(reqMethod, reqUrl string, queryString string, headers *(map[string]string)) error {
	sha512ReqPayload, _ := GetSHA512("")
	timestampStr := strconv.Itoa(int(time.Now().Unix()))
	payload := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", reqMethod, reqUrl, queryString, sha512ReqPayload, timestampStr)
	sign, _ := GetParamHmacSHA512Sign(gateio.secretKey, payload)
	(*headers)["Timestamp"] = timestampStr
	(*headers)["SIGN"] = sign
	(*headers)["KEY"] = gateio.accessKey
	return nil
}

func (gateio *Gateio) buildPostForm(reqMethod, path string, postForm *url.Values) error {
	postForm.Set("AccessKeyId", gateio.accessKey)
	postForm.Set("SignatureMethod", "HmacSHA256")
	postForm.Set("SignatureVersion", "2")
	postForm.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05"))
	domain := strings.Replace(gateio.baseUrl, "https://", "", len(gateio.baseUrl))
	payload := fmt.Sprintf("%s\n%s\n%s\n%s", reqMethod, domain, path, postForm.Encode())
	sign, _ := GetParamHmacSHA256Base64Sign(gateio.secretKey, payload)
	postForm.Set("Signature", sign)

	/**
	p, _ := pem.Decode([]byte(gateio.ECDSAPrivateKey))
	pri, _ := secp256k1_go.PrivKeyFromBytes(secp256k1_go.S256(), p.Bytes)
	signer, _ := pri.Sign([]byte(sign))
	signAsn, _ := asn1.Marshal(signer)
	priSign := base64.StdEncoding.EncodeToString(signAsn)
	postForm.Set("PrivateSignature", priSign)
	*/

	return nil
}

func (gateio *Gateio) toJson(params url.Values) string {
	parammap := make(map[string]string)
	for k, v := range params {
		parammap[k] = v[0]
	}
	jsonData, _ := json.Marshal(parammap)
	return string(jsonData)
}

func (gateio *Gateio) parseDepthData(tick map[string]interface{}, size int) *Depth {
	bids, _ := tick["bids"].([]interface{})
	asks, _ := tick["asks"].([]interface{})

	depth := new(Depth)
	n := 0
	for _, r := range asks {
		var dr DepthRecord
		rr := r.([]interface{})
		dr.Price = ToFloat64(rr[0])
		dr.Amount = ToFloat64(rr[1])
		depth.AskList = append(depth.AskList, dr)
		n++
		if n == size {
			break
		}
	}

	n = 0
	for _, r := range bids {
		var dr DepthRecord
		rr := r.([]interface{})
		dr.Price = ToFloat64(rr[0])
		dr.Amount = ToFloat64(rr[1])
		depth.BidList = append(depth.BidList, dr)
		n++
		if n == size {
			break
		}
	}

	sort.Sort(sort.Reverse(depth.AskList))

	return depth
}

func (gateio *Gateio) GetExchangeName() string {
	return GATEIO
}

func (gateio *Gateio) GetCurrenciesList() ([]string, error) {
	url := gateio.baseUrl + "/v1/common/currencys"

	ret, err := HttpGet(gateio.httpClient, url)
	if err != nil {
		return nil, err
	}

	data, ok := ret["data"].([]interface{})
	if !ok {
		return nil, errors.New("response format error")
	}
	fmt.Println(data)
	return nil, nil
}

func (gateio *Gateio) GetCurrenciesPrecision() ([]interface{}, error) {
	return nil, nil
}

func (gateio *Gateio) GetAllCurrencyPair() ([]CurrencyPair, error) {
	url := gateio.baseUrl + "/api/v4/spot/currency_pairs"

	ret, err := HttpGet3(gateio.httpClient, url, map[string]string{})
	if err != nil {
		return nil, err
	}

	var currencyPairs []CurrencyPair
	for _, v := range ret {
		_sym := v.(map[string]interface{})
		var currencyPair CurrencyPair
		base := _sym["base"].(string)
		quote := _sym["quote"].(string)
		if quote != "USDT" {
			continue
		}
		baseCurrency := NewCurrency(base, "")
		quoteCurrency := NewCurrency(quote, "")
		currencyPair = NewCurrencyPair(baseCurrency, quoteCurrency)
		currencyPairs = append(currencyPairs, currencyPair)
	}
	return currencyPairs, nil
}

func (gateio *Gateio) GetTimestamp() (int64, error) {
	url := gateio.baseUrl + "/v1/common/timestamp"
	ret, err := HttpGet(gateio.httpClient, url)
	if err != nil {
		return 0, err
	}
	data, ok := ret["data"].(int64)
	if !ok {
		return 0, err
	}
	return data, nil
}
