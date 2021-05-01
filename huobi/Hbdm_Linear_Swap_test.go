package huobi

import (
	"github.com/BTreeNewBee/goex"
	"net/http"
	"testing"
)

var linearSwap *HbdmLinearSwap

func init() {

	linearSwap = NewHbdmLinearSwap(&goex.APIConfig{
		HttpClient:   http.DefaultClient,
		Endpoint:     "https://api.btcgateway.pro",
		ApiKey:       "",
		ApiSecretKey: "",
		Lever:        5,
	})
}

func TestHbdmLinearSwap_GetAccountInfo(t *testing.T) {
	t.Log(linearSwap.GetFutureUserinfo())
}
