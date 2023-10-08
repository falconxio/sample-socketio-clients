//go:build sample_client
// +build sample_client

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
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

type StreamResponse struct {
	Status    string                `json:"status" default:"error"`
	Event     string                `json:"event"`
	RequestId string                `json:"request_id"`
	Body      []PriceUpdateResponse `json:"body,omitempty"`
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
		url = fmt.Sprintf("%s://%s", "wss", fws.Host)
	} else {
		url = fmt.Sprintf("%s://%s", "ws", fws.Host)
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
		"secret_key": fws.secret,
		"sign":       signature,
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

func (fws *FalconxWSClient) ReadMessages() {
	fws.readerActive = true
	defer func() { fws.readerActive = false }()
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
		if err == nil && data.Status == "error" {
			log.Print("Error received")

			// Prettifying the response
			s, _ := json.MarshalIndent(data, "", "\t")
			log.Println("Error Response: ", string(s))

			if data.Event == "auth_response" {
				fws.authenticated = false
				fws.authResponse <- fws.authenticated
			}

		} else {
			if data.Body == "Authentication successful" {
				fws.authenticated = true
				fws.authResponse <- fws.authenticated
			} else {
				var data1 StreamResponse

				err1 := json.Unmarshal(msg, &data1)
				if err1 == nil && fws.LogStreams {
					// Prettifying the response
					s, _ := json.MarshalIndent(data1, "", "\t")
					log.Println("Price Stream: ", string(s))
				}
			}
		}
	}
}

func safeRead(conn *websocket.Conn) (messageType int, p []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from ", r)
			messageType, p, err = -1, nil, errors.New("connection lost")
		}
	}()
	messageType, r, err := conn.NextReader()
	p, err = ioutil.ReadAll(r)
	return
}

func AuthenticateAndSubscribe(fxClient *FalconxWSClient) {
	isAuthenticated, err := fxClient.Authenticate()
	if !isAuthenticated {
		log.Fatal("Unable to authenticate. Err: ", err)
	} else {
		log.Println("Trying to sub")
		success, err := fxClient.Subscribe("ETH", "USD", "fx_ws_06102023", []float64{0.1, 1}, "ETH")
		if !success {
			log.Fatal("Unable to Subscribe. Err: ", err)
		}
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
	// fxClient.EnableSSL()
	fxClient.EnableRetry(1, nil)

	fxClient.Connect()

	AuthenticateAndSubscribe(fxClient)

	<-waitChan
}
