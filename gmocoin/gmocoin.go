package gmocoin

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.coin.z.com/"

type CoinAPIClient struct {
	key        string
	secret     string
	httpClient *http.Client
}

func NewCoinAPIClient(key, secret string) *CoinAPIClient {
	apiClient := &CoinAPIClient{key, secret, &http.Client{}}
	return apiClient
}

func (api CoinAPIClient) header(method, path, reqBody string) map[string]string {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)

	text := timestamp + method + path + reqBody
	hc := hmac.New(sha256.New, []byte(api.secret))
	hc.Write([]byte(text))
	sign := hex.EncodeToString(hc.Sum(nil))

	headers := map[string]string{
		"API-KEY":       api.key,
		"API-TIMESTAMP": timestamp,
		"API-SIGN":      sign,
	}
	return headers
}

func (api *CoinAPIClient) Request(method, path, apiType string, reqBody string) (string, error) {
	endpoint := baseURL + apiType + path
	headers := api.header(method, path, reqBody)

	req, err := http.NewRequest(method, endpoint, bytes.NewBufferString(reqBody))
	if err != nil {
		log.Printf("failed to create HTTP request: %v", err)
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := api.httpClient.Do(req)
	if err != nil {
		log.Printf("failed to send HTTP request: %v", err)
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read response body: %v", err)
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, body, "", " ")
	if err != nil {
		log.Printf("failed to indent JSON response: %v", err)
		return "", fmt.Errorf("failed to indent JSON response: %v", err)
	}

	return buf.String(), nil
}

type Asset struct {
	Amount         string `json:"amount"`
	Available      string `json:"available"`
	ConversionRate string `json:"conversionRate"`
	Symbol         string `json:"symbol"`
}

type AssetsResponse struct {
	Status       int     `json:"status"`
	Data         []Asset `json:"data"`
	ResponseTime string  `json:"responsetime"`
}

func (api *CoinAPIClient) GetAssets() (*AssetsResponse, error) {
	path := "/v1/account/assets"
	method := "GET"
	apiType := "private"

	response, err := api.Request(method, path, apiType, "")
	if err != nil {
		return nil, err
	}

	var assetsResponse AssetsResponse
	err = json.Unmarshal([]byte(response), &assetsResponse)
	if err != nil {
		return nil, err
	}

	if assetsResponse.Status != 0 {
		return nil, errors.New(fmt.Sprintf("API returned an error status: %d", assetsResponse.Status))
	}

	return &assetsResponse, nil
}