package webserver

import (
	"assetmanagement/app/database"
	"assetmanagement/appconfig"
	"assetmanagement/convert"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
)

type AssetResponse struct {
	Symbol    string `json:"symbol"`
	AmountJPY int    `json:"amount_jpy"`
}

type TotalResponse struct {
	TotalJPY       int             `json:"total_jpy"`
	GMOCoinAssets  []AssetResponse `json:"gmocoin_assets"`
	BittradeAssets []AssetResponse `json:"bittrade_assets"`
}

func AssetsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("app/html/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ApiAssetsHandler(w http.ResponseWriter, r *http.Request) {
	data, err := convert.FetchDataSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := database.NewDatabase()
	if err != nil {
		return
	}

	prevData, err := db.GetLatestData()
	if err != nil {
		return
	}

	if data.Total_JPY != prevData.Total_JPY ||
		!reflect.DeepEqual(data.GMOCoinAssets_JPY, prevData.GMOCoinAssets_JPY) ||
		!reflect.DeepEqual(data.BittradeAssets_JPY, prevData.BittradeAssets_JPY) {
		err = db.SaveData(data)
		if err != nil {
			return
		}
	}

	gmocoinAssets := make([]AssetResponse, 0, len(data.GMOCoinAssets_JPY))
	for symbol, amount := range data.GMOCoinAssets_JPY {
		gmocoinAssets = append(gmocoinAssets, AssetResponse{
			Symbol:    symbol,
			AmountJPY: amount,
		})
	}

	bittradeAssets := make([]AssetResponse, 0, len(data.BittradeAssets_JPY))
	for symbol, amount := range data.BittradeAssets_JPY {
		bittradeAssets = append(bittradeAssets, AssetResponse{
			Symbol:    symbol,
			AmountJPY: amount,
		})
	}

	responseData := TotalResponse{
		TotalJPY:       data.Total_JPY,
		GMOCoinAssets:  gmocoinAssets,
		BittradeAssets: bittradeAssets,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func Start() error {
	http.HandleFunc("/", AssetsHandler)
	http.HandleFunc("/api/assets", ApiAssetsHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", appconfig.AppConfig.Port), nil)
}
