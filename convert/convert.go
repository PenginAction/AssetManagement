package convert

import (
	"assetmanagement/appconfig"
	"assetmanagement/bittrade"
	"assetmanagement/gmocoin"
	"encoding/json"
	"fmt"
	"log"

	// "log"
	"strconv"
	"time"
)

type AssetSummary struct {
	Time               time.Time      `json:"time"`
	GMOCoinAssets_JPY  map[string]int `json:"gmocoin_assets_jpy"`
	BittradeAssets_JPY map[string]int `json:"bittrade_assets_jpy"`
	Total_JPY          int            `json:"total_jpy"`
}

// func (a *AssetSummary) AddAssetsJPY(symbol string, value int) {
// 	a.Assets_JPY[symbol] = value
// 	a.Total_JPY += value
// }

func GetAssetsJPY(symbol string, assets []gmocoin.Asset) int {
	for _, asset := range assets {
		if asset.Symbol == symbol {
			amount, _ := strconv.ParseFloat(asset.Amount, 64)
			conversionRate, _ := strconv.ParseFloat(asset.ConversionRate, 64)
			return int(amount * conversionRate)
		}
	}
	return 0
}

func getConversionRate(bitClient *bittrade.Client, symbol string) (float64, error) {
	tickers, err := bitClient.GetMarketTickers()
	if err != nil {
		return 0, err
	}

	for _, ticker := range tickers.Data {
		if ticker.Symbol == symbol+"jpy" {
			return ticker.Close, nil
		}
	}

	return 0, fmt.Errorf("conversion rate not found for symbol: %s", symbol)
}

func GetAssetsJPYBittrade(symbol string, assetsBittrade *bittrade.AssetsResponse, bitClient *bittrade.Client) int {
	for _, list := range assetsBittrade.Data.List {
		if list.Currency == symbol {
			conversionRate, _ := getConversionRate(bitClient, symbol)
			balance, _ := strconv.ParseFloat(list.Balance, 64)
			return int(balance * conversionRate)
		}
	}
	return 0
}

func FetchDataGMOCoin() ([]gmocoin.Asset, error) {
	apiClient := gmocoin.NewCoinAPIClient(appconfig.AppConfig.GmoapiKey, appconfig.AppConfig.GmoapiSecret)
	assetsResponse, _ := apiClient.GetAssets()

	var Assets []gmocoin.Asset
	for _, asset := range assetsResponse.Data {
		amount, _ := strconv.ParseFloat(asset.Amount, 64)

		if amount > 0 {
			Assets = append(Assets, asset)
		}
	}

	return Assets, nil
}

func FetchDataBittrade() (*bittrade.AssetsResponse, error) {
	apiClient := bittrade.NewClient(appconfig.AppConfig.BitapiKey, appconfig.AppConfig.BitapiSecret)
	assetsResponse, _ := apiClient.GetUserBalance()

	if assetsResponse.Status == "ok" {
		return assetsResponse, nil
	}

	return nil, fmt.Errorf("failed to fetch data from BitTrade")
}

func FetchDataSummary() (AssetSummary, error) {
	assetsGmocoin, _ := FetchDataGMOCoin()
	assetsBittrade, _ := FetchDataBittrade()

	var total_jpy int
	gmoassets := make(map[string]int)
	bitassets := make(map[string]int)

	for _, symbol := range appconfig.AppConfig.GmoCoinSymbols {
		jpy := GetAssetsJPY(symbol, assetsGmocoin)
		gmoassets[symbol] = jpy
		total_jpy += jpy
	}

	apiClientBittrade := bittrade.NewClient(appconfig.AppConfig.BitapiKey, appconfig.AppConfig.BitapiSecret)

	for _, symbol := range appconfig.AppConfig.BittradeSymbols {
		jpy := GetAssetsJPYBittrade(symbol, assetsBittrade, apiClientBittrade)
		bitassets[symbol] = jpy
		total_jpy += jpy
	}

	summary := AssetSummary{
		Time: time.Now(),
		// Assets_JPY: assets,
		GMOCoinAssets_JPY:  gmoassets,
		BittradeAssets_JPY: bitassets,
		Total_JPY:          total_jpy,
	}

	summaryJSON, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal summary: %v", err)
	}

	log.Println(string(summaryJSON))

	return summary, nil
}
