package bittrade

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const API_URL = "https://api-cloud.bittrade.co.jp"

type Client struct {
	AccessKey string
	SecretKey string
}

func NewClient(accessKey, secretKey string) *Client {
	return &Client{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
}

// signメソッドはペイロードを署名
func (c *Client) sign(payload string) string {
	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(payload))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// sendRequestメソッドは指定したパスにリクエストを送信
func (c *Client) sendRequest(path string) (string, error) {
	params := map[string]string{
		"AccessKeyId":      c.AccessKey,
		"SignatureMethod":  "HmacSHA256",
		"SignatureVersion": "2",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05"),
	}

	// キーを昇順にソート
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// ソートされたクエリ文字列を作成
	var queryString string
	for _, k := range keys {
		queryString += fmt.Sprintf("%s=%s&", k, url.QueryEscape(params[k]))
	}
	queryString = strings.TrimSuffix(queryString, "&")

	// 署名のペイロードを作成
	payload := fmt.Sprintf("GET\napi-cloud.bittrade.co.jp\n%s\n%s", path, queryString)

	// 署名を作成します。
	signature := c.sign(payload)

	// 署名をクエリ文字列に追加
	queryString += fmt.Sprintf("&Signature=%s", url.QueryEscape(signature))

	// リクエストを送信
	resp, err := http.Get(fmt.Sprintf("%s%s?%s", API_URL, path, queryString))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

type Account struct {
	Status string `json:"status"`
	Data   []struct {
		ID      int    `json:"id"`
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
		State   string `json:"state"`
	} `json:"data"`
}

func (c *Client) GetUserAccount() (*Account, error) {
	resp, err := c.sendRequest("/v1/account/accounts")
	if err != nil {
		return nil, err
	}

	var account Account
	err = json.Unmarshal([]byte(resp), &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

type AssetList struct {
	Currency string `json:"currency"`
	Type     string `json:"type"`
	Balance  string `json:"balance"`
}

type Asset struct {
	Id    int         `json:"id"`
	Type  string      `json:"type"`
	State string      `json:"state"`
	List  []AssetList `json:"list"`
}

type AssetsResponse struct {
	Status string `json:"status"`
	Data   Asset  `json:"data"`
}

func (c *Client) GetUserBalance() (*AssetsResponse, error) {
	account, err := c.GetUserAccount()
	if err != nil {
		return nil, err
	}

	if len(account.Data) == 0 {
		return nil, fmt.Errorf("no account data avaliable")
	}

	path := fmt.Sprintf("/v1/account/accounts/%d/balance", account.Data[0].ID)
	resp, err := c.sendRequest(path)
	if err != nil {
		return nil, err
	}

	var balance AssetsResponse
	err = json.Unmarshal([]byte(resp), &balance)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

type Ticker struct {
	Symbol   string  `json:"symbol"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Amount   float64 `json:"amount"`
	Vol      float64 `json:"vol"`
	Count    int     `json:"count"`
	Bid      float64 `json:"bid"`
	BidSize  float64 `json:"bidSize"`
	Ask      float64 `json:"ask"`
	AskSize  float64 `json:"askSize"`
}

type TickersResponse struct {
	Status string   `json:"status"`
	Data   []Ticker `json:"data"`
}

//全取引ペアの相場情報
func (c *Client) GetMarketTickers() (*TickersResponse, error) {
	resp, err := c.sendRequest("/market/tickers")
	if err != nil {
		return nil, err
	}

	var tickers TickersResponse
	err = json.Unmarshal([]byte(resp), &tickers)
	if err != nil {
		return nil, err
	}

	return &tickers, nil
}