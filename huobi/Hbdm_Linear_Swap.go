package huobi

import (
	"errors"
	"fmt"
	. "github.com/BTreeNewBee/goex"
	"github.com/BTreeNewBee/goex/internal/logger"
	"net/url"
	"time"
)

type HbdmLinearSwap struct {
	base *Hbdm
	c    *APIConfig
}

const (
	linearAccountApiPath      = "/linear-swap-api/v1/swap_account_info"
	linearCrossAccountApiPath = "/linear-swap-api/v1/swap_cross_account_info"
)

func NewHbdmLinearSwap(c *APIConfig) *HbdmLinearSwap {
	if c.Lever <= 0 {
		c.Lever = 10
	}

	return &HbdmLinearSwap{
		base: NewHbdm(c),
		c:    c,
	}
}

func (swap *HbdmLinearSwap) GetExchangeName() string {
	return HBDM_SWAP
}

func (swap *HbdmLinearSwap) GetFutureTicker(currencyPair CurrencyPair, contractType string) (*Ticker, error) {
	return nil, errors.New("not implement")
}

func (swap *HbdmLinearSwap) GetFutureDepth(currencyPair CurrencyPair, contractType string, size int) (*Depth, error) {
	return nil, errors.New("not implement")
}

func (swap *HbdmLinearSwap) GetFutureUserinfo(currencyPair ...CurrencyPair) (*FutureAccount, error) {
	var accountInfoResponse []struct {
		Symbol           string  `json:"symbol"`
		MarginBalance    float64 `json:"margin_balance"`
		MarginPosition   float64 `json:"margin_position"`
		MarginFrozen     float64 `json:"margin_frozen"`
		MarginAvailable  float64 `json:"margin_available"`
		ProfitReal       float64 `json:"profit_real"`
		ProfitUnreal     float64 `json:"profit_unreal"`
		RiskRate         float64 `json:"risk_rate"`
		LiquidationPrice float64 `json:"liquidation_price"`
	}

	param := url.Values{}
	if len(currencyPair) > 0 {
		param.Set("contract_code", currencyPair[0].ToSymbol("-"))
	}

	err := swap.base.doRequest(linearAccountApiPath, &param, &accountInfoResponse)
	if err != nil {
		return nil, err
	}

	var futureAccount FutureAccount
	futureAccount.FutureSubAccounts = make(map[Currency]FutureSubAccount, 4)

	for _, acc := range accountInfoResponse {
		currency := NewCurrency(acc.Symbol, "")
		futureAccount.FutureSubAccounts[currency] = FutureSubAccount{
			Currency:      currency,
			AccountRights: acc.MarginBalance,
			KeepDeposit:   acc.MarginPosition,
			ProfitReal:    acc.ProfitReal,
			ProfitUnreal:  acc.ProfitUnreal,
			RiskRate:      acc.RiskRate,
		}
	}

	err = swap.base.doRequest(linearCrossAccountApiPath, &url.Values{}, &accountInfoResponse)
	if err != nil {
		return nil, err
	}

	for _, acc := range accountInfoResponse {
		currency := NewCurrency("USDT", "")
		futureAccount.FutureSubAccounts[currency] = FutureSubAccount{
			Currency:      currency,
			AccountRights: acc.MarginBalance,
			KeepDeposit:   acc.MarginPosition,
			ProfitReal:    acc.ProfitReal,
			ProfitUnreal:  acc.ProfitUnreal,
			RiskRate:      acc.RiskRate,
		}
	}

	return &futureAccount, nil
}

func (swap *HbdmLinearSwap) PlaceFutureOrder(currencyPair CurrencyPair, contractType, price, amount string, openType, matchPrice int, leverRate float64) (string, error) {
	param := url.Values{}
	param.Set("contract_code", currencyPair.ToSymbol("-"))
	param.Set("client_order_id", fmt.Sprint(time.Now().UnixNano()))
	param.Set("price", price)
	param.Set("volume", amount)
	param.Set("lever_rate", fmt.Sprintf("%.0f", leverRate))

	direction, offset := swap.base.adaptOpenType(openType)
	param.Set("direction", direction)
	param.Set("offset", offset)
	logger.Info(direction, offset)

	if matchPrice == 1 {
		param.Set("order_price_type", "opponent")
	} else {
		param.Set("order_price_type", "limit")
	}

	var orderResponse struct {
		OrderId       string `json:"order_id_str"`
		ClientOrderId int64  `json:"client_order_id"`
	}

	err := swap.base.doRequest(placeOrderApiPath, &param, &orderResponse)
	if err != nil {
		return "", err
	}

	return orderResponse.OrderId, nil
}

func (swap *HbdmLinearSwap) LimitFuturesOrder(currencyPair CurrencyPair, contractType, price, amount string, openType int, opt ...LimitOrderOptionalParameter) (*FutureOrder, error) {
	orderId, err := swap.PlaceFutureOrder(currencyPair, contractType, price, amount, openType, 0, swap.c.Lever)
	return &FutureOrder{
		Currency:     currencyPair,
		OrderID2:     orderId,
		Amount:       ToFloat64(amount),
		Price:        ToFloat64(price),
		OType:        openType,
		ContractName: contractType,
	}, err
}

func (swap *HbdmLinearSwap) MarketFuturesOrder(currencyPair CurrencyPair, contractType, amount string, openType int) (*FutureOrder, error) {
	orderId, err := swap.PlaceFutureOrder(currencyPair, contractType, "", amount, openType, 1, 10)
	return &FutureOrder{
		Currency:     currencyPair,
		OrderID2:     orderId,
		Amount:       ToFloat64(amount),
		OType:        openType,
		ContractName: contractType,
	}, err
}

func (swap *HbdmLinearSwap) FutureCancelOrder(currencyPair CurrencyPair, contractType, orderId string) (bool, error) {
	param := url.Values{}
	param.Set("order_id", orderId)
	param.Set("contract_code", currencyPair.ToSymbol("-"))

	var cancelResponse struct {
		Errors []struct {
			ErrMsg    string `json:"err_msg"`
			Successes string `json:"successes,omitempty"`
		} `json:"errors"`
	}

	err := swap.base.doRequest(cancelOrderApiPath, &param, &cancelResponse)
	if err != nil {
		return false, err
	}

	if len(cancelResponse.Errors) > 0 {
		return false, errors.New(cancelResponse.Errors[0].ErrMsg)
	}

	return true, nil
}

func (swap *HbdmLinearSwap) GetFuturePosition(currencyPair CurrencyPair, contractType string) ([]FuturePosition, error) {
	param := url.Values{}
	param.Set("contract_code", currencyPair.ToSymbol("-"))

	var (
		tempPositionMap  map[string]*FuturePosition
		futuresPositions []FuturePosition
		positionResponse []struct {
			Symbol         string
			ContractCode   string  `json:"contract_code"`
			Volume         float64 `json:"volume"`
			Available      float64 `json:"available"`
			CostOpen       float64 `json:"cost_open"`
			CostHold       float64 `json:"cost_hold"`
			ProfitUnreal   float64 `json:"profit_unreal"`
			ProfitRate     float64 `json:"profit_rate"`
			Profit         float64 `json:"profit"`
			PositionMargin float64 `json:"position_margin"`
			LeverRate      float64 `json:"lever_rate"`
			Direction      string  `json:"direction"`
		}
	)

	err := swap.base.doRequest(getPositionApiPath, &param, &positionResponse)
	if err != nil {
		return nil, err
	}

	futuresPositions = make([]FuturePosition, 0, 2)
	tempPositionMap = make(map[string]*FuturePosition, 2)

	for _, pos := range positionResponse {
		if tempPositionMap[pos.ContractCode] == nil {
			tempPositionMap[pos.ContractCode] = new(FuturePosition)
		}
		switch pos.Direction {
		case "sell":
			tempPositionMap[pos.ContractCode].ContractType = pos.ContractCode
			tempPositionMap[pos.ContractCode].Symbol = NewCurrencyPair3(pos.ContractCode, "-")
			tempPositionMap[pos.ContractCode].SellAmount = pos.Volume
			tempPositionMap[pos.ContractCode].SellAvailable = pos.Available
			tempPositionMap[pos.ContractCode].SellPriceAvg = pos.CostOpen
			tempPositionMap[pos.ContractCode].SellPriceCost = pos.CostHold
			tempPositionMap[pos.ContractCode].SellProfitReal = pos.ProfitRate
			tempPositionMap[pos.ContractCode].SellProfit = pos.Profit
		case "buy":
			tempPositionMap[pos.ContractCode].ContractType = pos.ContractCode
			tempPositionMap[pos.ContractCode].Symbol = NewCurrencyPair3(pos.ContractCode, "-")
			tempPositionMap[pos.ContractCode].BuyAmount = pos.Volume
			tempPositionMap[pos.ContractCode].BuyAvailable = pos.Available
			tempPositionMap[pos.ContractCode].BuyPriceAvg = pos.CostOpen
			tempPositionMap[pos.ContractCode].BuyPriceCost = pos.CostHold
			tempPositionMap[pos.ContractCode].BuyProfitReal = pos.ProfitRate
			tempPositionMap[pos.ContractCode].BuyProfit = pos.Profit
		}
	}

	for _, pos := range tempPositionMap {
		futuresPositions = append(futuresPositions, *pos)
	}

	return futuresPositions, nil
}

func (swap *HbdmLinearSwap) GetFutureOrders(orderIds []string, currencyPair CurrencyPair, contractType string) ([]FutureOrder, error) {
	return nil, nil
}

func (swap *HbdmLinearSwap) GetFutureOrder(orderId string, currencyPair CurrencyPair, contractType string) (*FutureOrder, error) {
	var (
		orderInfoResponse []OrderInfo
		param             = url.Values{}
	)

	param.Set("contract_code", currencyPair.ToSymbol("-"))
	param.Set("order_id", orderId)

	err := swap.base.doRequest(getOrderInfoApiPath, &param, &orderInfoResponse)
	if err != nil {
		return nil, err
	}

	if len(orderInfoResponse) == 0 {
		return nil, errors.New("not found")
	}

	orderInfo := orderInfoResponse[0]

	return &FutureOrder{
		Currency:     currencyPair,
		ClientOid:    fmt.Sprint(orderInfo.ClientOrderId),
		OrderID2:     fmt.Sprint(orderInfo.OrderId),
		Price:        orderInfo.Price,
		Amount:       orderInfo.Volume,
		AvgPrice:     orderInfo.TradeAvgPrice,
		DealAmount:   orderInfo.TradeVolume,
		OrderID:      orderInfo.OrderId,
		Status:       swap.base.adaptOrderStatus(orderInfo.Status),
		OType:        swap.base.adaptOffsetDirectionToOpenType(orderInfo.Offset, orderInfo.Direction),
		LeverRate:    orderInfo.LeverRate,
		Fee:          orderInfo.Fee,
		ContractName: orderInfo.ContractCode,
		OrderTime:    orderInfo.CreatedAt,
	}, nil
}

func (swap *HbdmLinearSwap) GetFutureOrderHistory(pair CurrencyPair, contractType string, optional ...OptionalParameter) ([]FutureOrder, error) {
	params := url.Values{}
	params.Add("status", "0")     //all
	params.Add("type", "1")       //all
	params.Add("trade_type", "0") //all

	if contractType == "" || contractType == SWAP_CONTRACT {
		params.Add("contract_code", pair.AdaptUsdtToUsd().ToSymbol("-"))
	} else {
		return nil, errors.New("contract type is error")
	}

	MergeOptionalParameter(&params, optional...)

	var historyOrderResp struct {
		Orders     []OrderInfo `json:"orders"`
		RemainSize int64       `json:"remain_size"`
		NextId     int64       `json:"next_id"`
	}

	err := swap.base.doRequest(getHistoryOrderPath, &params, &historyOrderResp)
	if err != nil {
		return nil, err
	}

	var historyOrders []FutureOrder

	for _, ord := range historyOrderResp.Orders {
		historyOrders = append(historyOrders, FutureOrder{
			OrderID:      ord.OrderId,
			OrderID2:     fmt.Sprintf("%d", ord.OrderId),
			Price:        ord.Price,
			Amount:       ord.Volume,
			AvgPrice:     ord.TradeAvgPrice,
			DealAmount:   ord.TradeVolume,
			OrderTime:    ord.CreateDate,
			Status:       swap.base.adaptOrderStatus(ord.Status),
			Currency:     pair,
			OType:        swap.base.adaptOffsetDirectionToOpenType(ord.Offset, ord.Direction),
			LeverRate:    ord.LeverRate,
			Fee:          ord.Fee,
			ContractName: ord.ContractCode,
		})
	}

	return historyOrders, nil
}

func (swap *HbdmLinearSwap) GetUnfinishFutureOrders(currencyPair CurrencyPair, contractType string) ([]FutureOrder, error) {
	param := url.Values{}
	param.Set("contract_code", currencyPair.ToSymbol("-"))
	param.Set("page_size", "50")

	var openOrderResponse struct {
		Orders []OrderInfo
	}

	err := swap.base.doRequest(getOpenOrdersApiPath, &param, &openOrderResponse)
	if err != nil {
		return nil, err
	}

	openOrders := make([]FutureOrder, 0, len(openOrderResponse.Orders))
	for _, ord := range openOrderResponse.Orders {
		openOrders = append(openOrders, FutureOrder{
			Currency:   currencyPair,
			ClientOid:  fmt.Sprint(ord.ClientOrderId),
			OrderID2:   fmt.Sprint(ord.OrderId),
			Price:      ord.Price,
			Amount:     ord.Volume,
			AvgPrice:   ord.TradeAvgPrice,
			DealAmount: ord.TradeVolume,
			OrderID:    ord.OrderId,
			Status:     swap.base.adaptOrderStatus(ord.Status),
			OType:      swap.base.adaptOffsetDirectionToOpenType(ord.Offset, ord.Direction),
			LeverRate:  ord.LeverRate,
			Fee:        ord.Fee,
			OrderTime:  ord.CreatedAt,
		})
	}

	return openOrders, nil
}

func (swap *HbdmLinearSwap) GetContractValue(currencyPair CurrencyPair) (float64, error) {
	switch currencyPair {
	case BTC_USD, BTC_USDT:
		return 100, nil
	default:
		return 0, nil
	}
}

func (swap *HbdmLinearSwap) GetKlineRecords(contractType string, currency CurrencyPair, period KlinePeriod, size int, opt ...OptionalParameter) ([]FutureKline, error) {
	panic("not implement")
}

func (swap *HbdmLinearSwap) GetTrades(contractType string, currencyPair CurrencyPair, since int64) ([]Trade, error) {
	panic("not implement")
}

func (swap *HbdmLinearSwap) GetFee() (float64, error) {
	panic("not implement")
}

func (swap *HbdmLinearSwap) GetFutureIndex(currencyPair CurrencyPair) (float64, error) {
	panic("not implement")
}

func (swap *HbdmLinearSwap) GetDeliveryTime() (int, int, int, int) {
	panic("not implement")
}

func (swap *HbdmLinearSwap) GetFutureEstimatedPrice(currencyPair CurrencyPair) (float64, error) {
	panic("not implement")
}
