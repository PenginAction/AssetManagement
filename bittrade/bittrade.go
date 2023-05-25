package bittrade

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const baseURL = "https://api-cloud.bittrade.co.jp"

type BitTradeAPIClient struct {
	key        string
	secret     string
	httpClient *http.Client
}

func NewBitTradeAPIClient(key, secret string) *BitTradeAPIClient{
	apiClient := &BitTradeAPIClient{key, secret, &http.Client{}}
	log.Println("New BitTradeAPIClient created.")
	return apiClient
}

func (api BitTradeAPIClient) header(method, endpoint string, params map[string]string) map[string]string {
	// パラメータをソート．
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	log.Println("Parameters sorted.")

	// クエリ文字列を作成．
	query := ""
	for _, k := range keys {
		query += fmt.Sprintf("%s=%s&", url.QueryEscape(k), url.QueryEscape(params[k]))
	}
	// 末尾の'&'を削除
	query = strings.TrimSuffix(query, "&")
	log.Println("Query string created: ", query)

	// 署名する文字列を作成する．
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05")
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s", method, endpoint, query, timestamp)

	log.Println("String to sign created: ", stringToSign)

	//文字列の署名
	sign := api.sign(stringToSign)
	log.Println("Signature created: ", sign)

	headers := map[string]string{
		"Content-Type":     "application/json",
		"ACCESS-KEY":       api.key,
		"SIGNATURE-METHOD": "HmacSHA256",
		"SIGNATURE-VERSION": "2",
		"TIMESTAMP":        timestamp,
		"SIGNATURE":        sign,
	}
	log.Println("Headers created: ", headers)
	return headers
}


func (api *BitTradeAPIClient) sign(message string) string {
	h := hmac.New(sha256.New, []byte(api.secret))
	h.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	log.Println("Message signed: ", signature)
	return url.QueryEscape(signature)
}

func (api *BitTradeAPIClient) Request(method, endpoint string, params map[string]string) (string, error){
	headers := api.header(method, endpoint, params)

	reqBody := bytes.NewBufferString("")
	if method == "POST"{
		jsonValue, _ := json.Marshal(params)
		reqBody = bytes.NewBuffer(jsonValue)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, reqBody)
	if err != nil{
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	for k, v := range headers{
		req.Header.Set(k, v)
	}

	log.Println("Request headers: ", req.Header)

	res, err := api.httpClient.Do(req)
	if err != nil{
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer res.Body.Close()

	log.Printf("Response status: %s", res.Status)
	bodyBytes, _ := ioutil.ReadAll(res.Body)
	bodyString := string(bodyBytes)
	log.Printf("Response body: %s", bodyString)

	var buf bytes.Buffer
	err = json.Indent(&buf, bodyBytes, "", " ")
	if err != nil{
		return "", fmt.Errorf("failed to indent JSON response: %v", err)
	}

	return buf.String(), nil
}


type UserInfoData struct{
	ID     int     `json:"id"`
	Type   string  `json:"type"`
	State  string  `json:"state"`
	UserID int     `json:"user-id"`
}

type UserInfoResponse struct{
	Status string     `json:"status"`
	Data   UserInfoData `json:"data"`
}

func (api *BitTradeAPIClient) GetAccountInfo() (*UserInfoResponse, error){
	method := "GET"
	path := "/v1/account/accounts"
	params := map[string]string{}

	resp, err := api.Request(method, path, params)
	if err != nil{
		return nil, err
	}

	var userinfoResponse UserInfoResponse
	err = json.Unmarshal([]byte(resp), &userinfoResponse)
	if err != nil{
		return nil, err
	}

	if userinfoResponse.Status != "ok"{
		return nil, fmt.Errorf("API returned an error status: %s", userinfoResponse.Status)
	}
	
	return &userinfoResponse, nil
}


type Asset struct {
	Currency string `json:"currency"`
	Type     string `json:"type"`
	Balance  string `json:"balance"`
}

type AssetsData struct {
	ID     int     `json:"id"`
	Type   string  `json:"type"`
	State  string  `json:"state"`
	List   []Asset `json:"list"`
	UserID int     `json:"user-id"`
}

type AssetsResponse struct {
	Status string     `json:"status"`
	Data   AssetsData `json:"data"`
}

func (api *BitTradeAPIClient) GetAssets() (*AssetsResponse, error){
	userInfo, err := api.GetAccountInfo()
	if err != nil{
		return nil, err
	}

	accountID := userInfo.Data.ID
	path := fmt.Sprintf("/v1/account/accounts/%d/balance", accountID)
	method := "GET"
	params := make(map[string]string)

	resp, err := api.Request(method, path, params)
	if err != nil{
		return nil, err
	}

	var assetsResponse AssetsResponse
	err = json.Unmarshal([]byte(resp), &assetsResponse)
	if err != nil{
		return nil, err
	}

	if assetsResponse.Status != "ok"{
		return nil, fmt.Errorf("API returned an error status: %s", assetsResponse.Status )
	}
	
	return &assetsResponse, nil
}