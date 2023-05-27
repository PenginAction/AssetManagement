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
		ID     int    `json:"id"`
		Type   string `json:"type"`
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
