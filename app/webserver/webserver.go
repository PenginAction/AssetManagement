package webserver

import (
	"assetmanagement/app/database"
	"assetmanagement/appconfig"
	"assetmanagement/convert"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type AssetResponse struct {
	Symbol    string `json:"symbol"`
	AmountJPY int    `json:"amount_jpy"`
}

type TotalResponse struct {
	TotalJPY int             `json:"total_jpy"`
	Assets   []AssetResponse `json:"assets"`
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
	if err != nil{
		return 
	}

	err = db.SaveData()
	if err != nil{
		return
	}

	responseData := TotalResponse{
		TotalJPY: data.Total_JPY,
		Assets: []AssetResponse{
			{
				Symbol:    appconfig.AppConfig.Symbol1,
				AmountJPY: data.BTC_JPY,
			},
			{
				Symbol:    appconfig.AppConfig.Symbol2,
				AmountJPY: data.XEM_JPY,
			},
			{
				Symbol:    appconfig.AppConfig.Symbol3,
				AmountJPY: data.ADA_JPY,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

func Start() error {
	http.HandleFunc("/", AssetsHandler)
	http.HandleFunc("/api/assets", ApiAssetsHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", appconfig.AppConfig.Port), nil)
}