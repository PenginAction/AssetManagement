package convert

import (
	"assetmanagement/appconfig"
	"assetmanagement/gmocoin"
	"assetmanagement/utils"
	"encoding/json"
	"log"
	"strconv"
	"time"
)

type AssetSummary struct{
	Time time.Time `json:"time"`
	BTC_JPY int `json:"btc_jpy"`
	XEM_JPY int `json:"xem_jpy"`
	ADA_JPY int `json:"ada_jpy"`
	Total_JPY int `json:"total_jpy"`
}

func GetAssetsJPY(symbol string, assets []gmocoin.Asset) int{
	for _, asset := range assets{
		if asset.Symbol == symbol{
			amount, _ := strconv.ParseFloat(asset.Amount, 64)
			conversionRate, _ := strconv.ParseFloat(asset.ConversionRate, 64)
			return int(amount * conversionRate)
		}
	}
	return 0
}

func FetchData() ([]gmocoin.Asset, error){
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

func FetchDataSummary() (AssetSummary, error) {
	assets, err := FetchData()
	if err != nil {
		return AssetSummary{}, err
	}

	btc_jpy := GetAssetsJPY(appconfig.AppConfig.Symbol1, assets)
	xem_jpy := GetAssetsJPY(appconfig.AppConfig.Symbol2, assets)
	ada_jpy := GetAssetsJPY(appconfig.AppConfig.Symbol3, assets)
	total_jpy := utils.Sum(btc_jpy, xem_jpy, ada_jpy)

	summary := AssetSummary{
		Time:      time.Now(),
		BTC_JPY:   btc_jpy,
		XEM_JPY:   xem_jpy,
		ADA_JPY:   ada_jpy,
		Total_JPY: total_jpy,
	}

	jsonData, err := json.MarshalIndent(summary, "", " ")
	if err != nil {
		log.Println("Error marshaling data", err)
		return AssetSummary{}, err
	}

	log.Println(string(jsonData))

	return summary, nil
}