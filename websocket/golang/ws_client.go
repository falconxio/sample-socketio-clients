//go:build ws_client
// +build ws_client

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type ResponseMessage struct {
	Status    string      `json:"status" default:"error"`
	Error     interface{} `json:"error,omitempty"`
	Event     string      `json:"event"`
	RequestId string      `json:"request_id"`
	Body      interface{} `json:"body,omitempty"`
}

type PriceUpdateResponse struct {
	TCreate       int64    `json:"t_create"`
	Quantity      float64  `json:"quantity"`
	QuantityToken string   `json:"quantity_token"`
	SellPrice     *float64 `json:"sell_price"`
	BuyPrice      *float64 `json:"buy_price"`
	BaseToken     string   `json:"base_token"`
	QuoteToken    string   `json:"quote_token"`
}

type QuantityLevels struct {
	Token  string    `json:"token"`
	Levels []float64 `json:"levels"`
}

type SubscribeRequest struct {
	RequestId  string         `json:"request_id"`
	Action     string         `json:"action"`
	BaseToken  string         `json:"base_token"`
	QuoteToken string         `json:"quote_token"`
	Quantity   QuantityLevels `json:"quantity"`
}

type UnSubscribeRequest struct {
	RequestId  string `json:"request_id"`
	Action     string `json:"action"`
	BaseToken  string `json:"base_token"`
	QuoteToken string `json:"quote_token"`
}

type DataRequest struct {
	RequestId   string `json:"request_id"`
	Action      string `json:"action"`
	RequestType string `json:"request_type"`
}

// Utility Functions:
func getSign(timeint int64, secretKey string, path string) string {
	timestamp := strconv.FormatInt(timeint, 10)
	method := "GET"

	message := strings.Join([]string{timestamp, method, path, ""}, "")
	signature := GetSignedMessageForKey([]byte(message), secretKey)
	return signature
}

func GetSignedMessageForKey(message []byte, secretKey string) string {
	key, _ := base64.StdEncoding.DecodeString(secretKey)
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

type FalconxWSClient struct {
	Host          string
	Path          string
	SSL           bool
	apiKey        string
	secret        string
	passphrase    string
	Conn          *websocket.Conn
	RetryOnError  bool
	NumRetries    *uint64
	RetryDelay    time.Duration
	retryCount    uint64
	authenticated bool
	LogStreams    bool
	readerActive  bool
	authResponse  chan bool
	interruptRead chan bool
}

func NewFalconxWSClient(host string, path string) *FalconxWSClient {
	return &FalconxWSClient{
		Host:          host,
		Path:          path,
		interruptRead: make(chan bool),
		authResponse:  make(chan bool),
		LogStreams:    true,
	}
}

func (fws *FalconxWSClient) DisableLogging() {
	fws.LogStreams = false
}

func (fws *FalconxWSClient) EnableSSL() {
	fws.SSL = true
}

func (fws *FalconxWSClient) SetAuth(apiKey string, secret string, passphrase string) {
	fws.apiKey = apiKey
	fws.secret = secret
	fws.passphrase = passphrase
}

func (fws *FalconxWSClient) EnableRetry(retryDelayInSeconds uint64, numOfRetries *uint64) {
	fws.RetryOnError = true
	if retryDelayInSeconds > 0 {
		fws.RetryDelay = time.Second * time.Duration(retryDelayInSeconds)
	} else {
		fws.RetryDelay = time.Second * 1
	}
	fws.NumRetries = numOfRetries
}

func (fws *FalconxWSClient) Connect() {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	var url string
	if fws.SSL {
		url = fmt.Sprintf("%s://%s%s", "wss", fws.Host, fws.Path)
	} else {
		url = fmt.Sprintf("%s://%s%s", "ws", fws.Host, fws.Path)
	}

	log.Println("Trying to connect to ", url)
	conn, a, err := dialer.Dial(url, nil)

	if err != nil {
		log.Print("Error connecting to WebSocket server:", a, err)
		if fws.RetryOnError && (fws.NumRetries == nil || fws.retryCount < *fws.NumRetries) {
			log.Printf("Retrying in %+v", fws.RetryDelay)
			time.Sleep(fws.RetryDelay)
			fws.Connect()
		}
		return
	} else {
		log.Printf("Connection successful")
	}

	fws.Conn = conn
	if !fws.readerActive {
		go fws.ReadMessages()
	}
	// defer conn.Close()
}

func (fws *FalconxWSClient) Authenticate() (bool, error) {
	timestamp := time.Now().Unix()

	signature := getSign(timestamp, fws.secret, fws.Path)

	req := map[string]interface{}{
		"action":     "auth",
		"api_key":    fws.apiKey,
		"passphrase": fws.passphrase,
		"signature":  signature,
		"timestamp":  timestamp,
		"request_id": "my_request",
	}
	err := fws.Conn.WriteJSON(req)

	if err != nil {
		log.Println("Error: ", err)
		return false, err
	}
	log.Println("Sent Auth msg")
	return <-fws.authResponse, nil
}

func (fws *FalconxWSClient) Subscribe(baseToken string, quoteToken string, clientRequestId string, levels []float64, quantityToken string) (bool, error) {
	req := SubscribeRequest{
		Action:     "subscribe",
		RequestId:  clientRequestId,
		BaseToken:  baseToken,
		QuoteToken: quoteToken,
		Quantity: QuantityLevels{
			Token:  quantityToken,
			Levels: levels,
		},
	}
	log.Println("Sending request -> ", req)
	err := fws.Conn.WriteJSON(req)
	if err != nil {
		log.Printf("error occurred while writing json %s", err)
		return false, err
	}

	return true, nil
}

func (fws *FalconxWSClient) UnSubscribe(baseToken string, quoteToken string, clientRequestId string) (bool, error) {
	req := UnSubscribeRequest{
		Action:     "unsubscribe",
		RequestId:  clientRequestId,
		BaseToken:  baseToken,
		QuoteToken: quoteToken,
	}
	log.Println("Sending request -> ", req)
	err := fws.Conn.WriteJSON(req)
	if err != nil {
		log.Printf("error occurred while writing json %s", err)
		return false, err
	}

	return true, nil
}

func (fws *FalconxWSClient) RequestData(requestType string, clientRequestId string) (bool, error) {

	req := DataRequest{
		Action:      "data_request",
		RequestId:   clientRequestId,
		RequestType: requestType,
	}
	log.Println("Sending request -> ", req)
	err := fws.Conn.WriteJSON(req)
	if err != nil {
		log.Printf("error occurred while writing json %s", err)
		return false, err
	}

	return true, nil
}

func (fws *FalconxWSClient) ReadMessages() {
	fws.readerActive = true
	defer func() { fws.readerActive = false }()
	// count := 0
	for {
		_, msg, errRead := safeRead(fws.Conn)
		if errRead != nil {
			fws.Conn.Close()
			// Try to reconnect
			fws.Connect()
			go AuthenticateAndSubscribe(fws)
		}

		var data ResponseMessage

		err := json.Unmarshal(msg, &data)
		switch data.Event {
		case "auth_response":
			{
				if err == nil && data.Status == "error" {
					log.Println("Authentication Failed. Err: ", data.Error, data.Body)
					fws.authenticated = false
				} else {
					log.Println("Authentication Successful", data.Body)
					fws.authenticated = true
				}
				fws.authResponse <- fws.authenticated
			}
		case "subscribe_response":
			{
				if err == nil && data.Status == "error" {
					log.Println("Unable to subscribe. Err: ", data.Error, data.Body)
				} else {
					log.Println("Subscription Successful. Response: ", data.Body)
				}
			}
		case "unsubscribe_response":
			{
				if err == nil && data.Status == "error" {
					log.Println("Unable to unsubscribe. Err: ", data.Error, data.Body)
				} else {
					log.Println("UnSubscription Successful. Response: ", data.Body)
				}
			}
		case "data_response":
			{
				if err == nil && data.Status == "error" {
					log.Println("Unable to fetch requested data. Err: ", data.Error, data.Body)
				} else {
					log.Println("Data Response: ", data.Body)
				}
			}
		case "stream":
			{
				streamData := data.Body.([]interface{})
				if fws.LogStreams {
					// Prettifying the response
					s, _ := json.MarshalIndent(streamData, "", "\t")
					log.Println("Price Stream: ", string(s))
				}
			}
		case "error_response":
			{
				log.Println("Error Response Received: ", data.Body)
			}
		}
	}
}

func safeRead(conn *websocket.Conn) (messageType int, p []byte, err error) {
	messageType, r, err := conn.NextReader()
	if err == nil {
		p, err = ioutil.ReadAll(r)
	}
	return
}

func AuthenticateAndSubscribe(fxClient *FalconxWSClient) {
	isAuthenticated, err := fxClient.Authenticate()
	if !isAuthenticated {
		log.Fatal("Unable to authenticate. Err: ", err)
	} else {
		log.Println("Trying to subscribe")
		success, err := fxClient.Subscribe("ETH", "USD", "fx_ws_06102023", []float64{0.1, 1}, "ETH")
		if !success {
			log.Fatal("Unable to Subscribe. Err: ", err)
		}

		// time.Sleep(5 * time.Second)
		// fxClient.UnSubscribe("ETH", "USD", "fx_ws_06102023")
	}
}

func AuthenticateAndRequestData(fxClient *FalconxWSClient) {
	isAuthenticated, err := fxClient.Authenticate()
	if !isAuthenticated {
		log.Fatal("Unable to authenticate. Err: ", err)
		fxClient.Conn.Close()
	} else {
		log.Println("Trying to subscribe")
		success, err := fxClient.RequestData("max_levels", "fx_ws_14102023")
		success, err = fxClient.RequestData("allowed_markets", "fx_ws_15102023")
		success, err = fxClient.RequestData("max_connections", "fx_ws_16102023")
		if !success {
			log.Fatal("Unable to Subscribe. Err: ", err)
		}

		// time.Sleep(5 * time.Second)
		// fxClient.UnSubscribe("ETH", "USD", "fx_ws_06102023")
	}
}

func main() {
	waitChan := make(chan bool)

	apiKey := "xxx"
	secret := "xxx"
	passphrase := "xxx"

	host := "stream.falconx.io"
	path := "/price.tickers"

	fxClient := NewFalconxWSClient(host, path)
	fxClient.SetAuth(apiKey, secret, passphrase)
	fxClient.EnableSSL()
	fxClient.EnableRetry(1, nil)

	fxClient.Connect()

	AuthenticateAndSubscribe(fxClient)
	// AuthenticateAndRequestData(fxClient)

	<-waitChan
}
