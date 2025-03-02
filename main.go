package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	market_coins_url = "https://api.coindcx.com/exchange/v1/markets"
	base_url         = "https://public.coindcx.com/market_data/orderbook?pair="
	trade_url        = "https://public.coindcx.com/market_data/trade_history?pair=I-"
)

var CoinWithINR = []string{"XAI", "SC", "PEIPEI", "GMT", "ZRO", "AVAX", "VINE", "1MBABYDOGE", "VVV", "OMNI", "ADA", "BENDOG", "POL", "FLOKI", "AZERO", "CAKE", "LRC", "FARTCOIN", "KAS", "DFI", "XCAD", "POPCAT", "AR", "CATS", "FLM", "HNT", "YFII", "INJ", "MIGGLES", "MYRIA", "ULTIMA", "CHZ", "POLYX", "WIF", "PEN", "TURBO", "SPX", "KAITO", "TOSHI", "METIS", "UMA", "STRAX", "FLT", "VINU", "DIONE", "MYRO", "AIXBT", "SUNDOG", "BSV", "BAN", "1000CHEEMS", "ACT", "MEME", "COOKIE", "AKT", "ALICE", "DEFI", "ALT", "ETC", "BERA", "BSW", "D", "AVAIL", "COW", "AXS", "SHIB", "ATH", "MANTA", "BRETT", "GRASS", "JASMY", "SOLV", "TST", "ALGO", "JUP", "BRISE", "PNUT", "CGPT", "ROOT", "HIPPO", "PONKE", "ONDO", "AERO", "TOMI", "AAVE", "DGB", "HBAR", "AIOZ", "TRB", "ICP", "FIL", "PHB", "LINK", "HYPE", "MOVE", "ARB", "VELODROME", "KNC", "MTL", "ACX", "DYM", "TFUEL", "ZEC", "SLERF", "FET", "CFG", "WAXL", "ELY", "ETHFI", "SILLY", "ILV", "EGLD", "LEVER", "LMWR", "ZIG", "STX", "SUSHI", "THE", "ENA", "LINA", "GOAT", "TNSR", "CSIX", "SAGA", "PDA", "PRCL", "LUNA", "RSR", "SYN", "NMR", "BDX", "PROS", "BCH", "NULS", "ZCX", "G", "S", "VRA", "VET", "LOOM", "ICE", "NIBI", "VSC", "SKY", "MBL", "FIDA", "TIA", "XVG", "CRV", "VELO", "WIN", "USUAL", "EWT", "SNX", "GRT", "USDC", "IMX", "OMG", "MOODENG", "SUN", "RENDER", "IOTX", "ARTFI", "WEN", "EMT", "TON", "XR", "LISTA", "X", "BONK", "BEAMX", "BIO", "COTI", "DMTR", "PUSH", "SUI", "BNB", "SPELL", "PYTH", "NEAR", "BIGTIME", "ME", "YFI", "NOT", "E4C", "MEMEFI", "CAT", "DEGEN", "PEPE", "VIRTUAL", "PHA", "XLM", "BAX", "BOME", "STRK", "BTTC", "TAO", "ZEREBRO", "GRIFFAIN", "W", "LAYER", "DOGS", "OM", "GOATS", "SEI", "VOLT", "SUPER", "PEOPLE", "NEIROETH", "GALA", "PENGU", "LDO", "TRUMP", "CHILLGUY", "XYO", "TRX", "ANIME", "APT", "DYDX", "HOT", "UNI", "GEOD", "WLD", "PAXG", "MELANIA", "AI16Z", "MERL", "SDEX", "MAJOR", "ALPACA", "HMSTR", "IQ", "NAKA", "PIXEL", "ARKM", "SKL", "BULL", "CREAM", "APE", "ACH", "WEMIX", "DATA", "ALEX", "STORJ", "QNT", "STC", "SYS", "RONIN", "NEO", "EIGEN", "LAT", "JOE", "MINA", "PORTAL", "DRIFT", "RIF", "ZIL", "AGG", "PERP", "AST", "SLP", "COPI", "OGN", "ORCA", "NKN", "KSM", "UFT", "CETUS", "BCUT", "ZETA", "SCR", "GAS", "VENOM", "HIGH", "PROM", "THETA", "SXP", "TLM", "COMP", "CELR", "GMMT", "ONE", "WAXP", "AMP", "LCX", "REQ", "FTT", "WAVES", "MASA", "RVN", "ZK", "DOT", "TKO", "IO", "DENT", "CFX", "MKR", "LUMIA", "EOS", "BANANA", "DODO", "PHIL", "MBOX", "OSMO", "CHR", "VANA", "TOKEN", "ROSE", "MEW", "CKB", "PYR", "XDC", "RLC", "CVC", "TWT", "BICO", "TUSD", "MAHA", "FTN", "SANTOS", "FUN", "XAUT", "BOBA", "FORT", "LPT", "MARSH", "ADX", "ONG", "TLOS", "IOST", "REN", "PDEX", "ORBS", "EURT", "ZBU", "SOLO", "QI", "CTC", "PRO", "VTHO", "XCN", "POWR", "TEL", "DAO", "AUDIO", "CSPR", "CTSI", "SIDUS", "CCD", "ADS", "XEM", "MASK", "NFP", "KAVA", "XDB", "CTK", "ELF", "MNT", "BLUR", "WOO", "ENS", "MAVIA", "OAS", "BB", "CELO", "ALPHA", "REZ", "ATOM", "RAY", "ORDI", "ANKR", "DIA", "XTZ", "GLM", "MOVR", "CORE", "DAI"} // Keep your full list

type OrderBook struct {
	Timestamp int64             `json:"timestamp"`
	Asks      map[string]string `json:"asks"`
	Bids      map[string]string `json:"bids"`
}

type TradeHistory struct {
	P float64 `json:"p"`
	Q float64 `json:"q"`
	S string  `json:"s"`
	T float64 `json:"T"`
	M bool    `json:"m"`
}

var (
	coinWithProfit 		= make(map[string]float64)
	profitMutex    		= &sync.Mutex{}
	httpClient     		= &http.Client{Timeout: 150 * time.Second}
	realCryptoShit      = make(map[string]float64)
	realCryptoShitMutex = &sync.Mutex{}
	seen 				  []string
	loop_iteration 		= 0
)

func main() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	go func ()  {
		for range ticker.C {
			seen = []string{}
			fmt.Println("Seen slice has been reset.")
		}
	}()

	for true {

		async_trade_history()
		time.Sleep(30 * time.Second)
		// fmt.Println(realCryptoShit)
		get_profit()
		time.Sleep(30 * time.Second)
		// fmt.Println(coinWithProfit)
		for coin, profit := range coinWithProfit {
			if !seen_list(coin) {
				seen = append(seen, coin)
				message := fmt.Sprintf("Coin Name:  %s \nTotal profit: %.2f \nLast Traded: %2.fs ago", coin, profit, realCryptoShit[coin] + 30.0)	
				fmt.Println(message)
				http.PostForm("https://api.pushover.net/1/messages.json", url.Values{
					"token":   {"a7y4swewmxje1xd4e7mk29wy6r5ged"},
					"user":    {"u964gnk8jyubzzoysrd8bnsorhd9nv"},
					"message": {message},
				})
			} 
		}
		
		for idx := range coinWithProfit {
			delete(coinWithProfit, idx)
		}
		
		for i := range realCryptoShit {
			delete(realCryptoShit, i)
		}
		
		loop_iteration += 1
		fmt.Println("Loop iteration: ", loop_iteration)
	}
}

func seen_list(coin string) bool {
	for _, s := range seen {
		if s == coin {
			return true
		}
	}
	return false
}

func get_profit() {
	var wg sync.WaitGroup
	for coin := range realCryptoShit {
		wg.Add(1)
		go func(c string) {
			defer func() {
				wg.Done()
			}()
			processCoin(c)
		}(coin)
	}

	wg.Wait()
}

func processCoin(coin string) {
	coinURL := base_url + "B-" + coin + "_INR"

	resp, err := httpClient.Get(coinURL) 
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", coin, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", coin, err)
		return
	}

	var orderbook OrderBook
	if err := json.Unmarshal(body, &orderbook); err != nil {
		fmt.Printf("Error parsing %s: %v\n", coin, err)
		return
	}
	// Get best prices
	bestAsk, bestBid := getBestPrices(orderbook)
	if bestAsk == 0 || bestBid == 0 {
		return
	}

	investment := 2000.0
	profit := (investment/bestBid)*bestAsk - investment - 30
	if profit > 60.0 {
		profitMutex.Lock()
		coinWithProfit[coin] = profit
		profitMutex.Unlock()
	}
}

func getBestPrices(orderbook OrderBook) (float64, float64) {
	var askPrices []float64
	for priceStr := range orderbook.Asks {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			continue
		}
		askPrices = append(askPrices, price)
	}
	if len(askPrices) == 0 {
		return 0, 0
	}
	sort.Float64s(askPrices)
	bestAsk := askPrices[0]

	var bidPrices []float64
	for priceStr := range orderbook.Bids {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			continue
		}
		bidPrices = append(bidPrices, price)
	}
	if len(bidPrices) == 0 {
		return 0, 0
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(bidPrices)))
	bestBid := bidPrices[0]

	return bestAsk, bestBid
}

func async_trade_history() {
	var wg sync.WaitGroup

	for _, coin := range CoinWithINR {
		wg.Add(1)

		go func(c string) {
			defer func() {
				wg.Done()
			}()
			processTradeHistory(c)
		}(coin)
	}
	wg.Wait()
}

func processTradeHistory(coin string) {
	coinUrl := trade_url + coin + "_INR" + "&limit=10"

	resp, err := httpClient.Get(coinUrl) 
	if err != nil {
		fmt.Printf("Error fetching trade history for %s: %v\n", coin, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP error for %s: Status code: %d\n", coin, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading trade history body for %s: %v\n", coin, err)
		return
	}

	var tradehistory []TradeHistory
	if err := json.Unmarshal(body, &tradehistory); err != nil {
		fmt.Printf("Error parsing trade history JSON for %s: %v\n", coin, err)
		fmt.Printf("Response body was: %s\n", string(body)) // Print the response body for debugging
		return
	}

	if len(tradehistory) == 0 {
		fmt.Printf("No trade history data received for %s\n", coin)
		return
	}

	var times []float64
	for _, trade := range tradehistory {
		times = append(times, trade.T)
	}

	sort.Sort(sort.Reverse(sort.Float64Slice(times)))

	var (
		diff      float64
		timeNow   int64
		j         = 1
	)
	// var highTimeDiff []float64
	timeNow = time.Now().UnixMilli() / 1000
	for idx := range 9 {
		if j >= len(times) { 
			break 
		}
		T1 := times[idx] / 1000
		T2 := times[j] / 1000
		diff = T1 - T2
		recentDiff := (float64(timeNow)) - (times[0] / 1000)
		// fmt.Println(diff, recentDiff)
		// fmt.Printf("Diff: %.2f Recent Diff: %.2f \n", diff, recentDiff)
		if diff > 30 || recentDiff > 30.0 {
			// highTimeDiff = append(highTimeDiff, diff)
			return
		}
		j += 1
		// fmt.Println(highTimeDiff)
	}
	// fmt.Println(coin, condition)
	for _, trade := range tradehistory {
		T1 := (float64(timeNow) * 1000) / 1000
		T2 := times[0] / 1000
		// formatted_T1 := strconv.FormatFloat(T1, 'f', 6, 64)
		// formatted_T2 := strconv.FormatFloat(T2, 'f', 6, 64)
		// parsed_T1, _ := strconv.ParseFloat(formatted_T1, 64)
		// parsed_T2, _ := strconv.ParseFloat(formatted_T2, 64)
		realCryptoShitMutex.Lock()
		removedINRpart := strings.TrimSuffix(trade.S, "INR")
		realCryptoShit[removedINRpart] = T1 - T2
		realCryptoShitMutex.Unlock()
		break
	}
	
}