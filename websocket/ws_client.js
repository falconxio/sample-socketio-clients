const WebSocket = require('ws');
const CryptoJS = require('crypto-js'); // crypto-js@4.1.1


class FXClient {
  constructor(url, path, ssl, apiKey, passphrase, secretKey, onConnectCallback) {
    this.url = url;
    this.path = path;
    this.ssl = ssl;
    this.onConnectCallback = onConnectCallback;
    this.apiKey = apiKey;
    this.passphrase = passphrase;
    this.secretKey = secretKey;
    this.connection = null;
  }

  createAuthRequest() {
    const timestamp = parseInt(new Date().getTime() / 1000);
    const message = timestamp + 'GET' + this.path;
    const hmacKey = CryptoJS.enc.Base64.parse(this.secretKey);
    const signature = CryptoJS.HmacSHA256(message, hmacKey);
    const signatureB64 = signature.toString(CryptoJS.enc.Base64);
    return {
      'action': 'auth',
      'request_id': "my_auth_request_0",
      'sign': signatureB64,
      'timestamp': timestamp,
      'api_key': this.apiKey,
      'passphrase': this.passphrase,
    }
  }

  onConnect() {
    console.log(`${new Date()} Connected to websocket server`);
    this.authenticate()
  }

  authenticate() {
    const authRequest = this.createAuthRequest()
    this.connection.send(JSON.stringify(authRequest))
  }

  onError(msg) {
    console.log(`Error. Message received:`, msg)
  }

  onClose(msg) {
    console.log(`${new Date()}Lol. Close:`, msg)
  }

  connect() {
    if (this.connection) {
    }
    const protocol = this.ssl ? "wss" : "ws"
    const finalUrl = `${protocol}://${this.url}${this.path}`
    console.log(finalUrl)
    this.connection = new WebSocket(finalUrl);
    this.connection.on('open', this.onConnect.bind(this));
    this.connection.on('close', this.onClose.bind(this));
    this.connection.on('end', this.onClose.bind(this));
    this.connection.on('error', this.onError.bind(this));
    this.connection.on('message', this.onMessage.bind(this));
    this.connection.onclose = function (ev) {
      console.info(`${new Date()}Websocket closed.`);
    };
  }

  onMessage(msg) {
    const stringData = String.fromCharCode.apply(null, msg)
    const jsonMessage = JSON.parse(stringData)
    switch(jsonMessage.event) {
      case "auth_response": {
        if(jsonMessage.status == "success"){
          // Subscribe
          // this.subscribe()

          // Fetch data
          this.fetchData("max_levels")
          this.fetchData("allowed_markets")
          this.fetchData("max_connections")
        }else{
          console.log("Authentication Failure")
        }
        break
      }
      case "subscribe_response": {
        if(jsonMessage.status == "success"){
          console.log("Subscription successful  -> ", jsonMessage)
        }else{
          console.log("Subscription Failed -> ", jsonMessage)
        }
        break
      }
      case "unsubscribe_response": {
        if(jsonMessage.status == "success"){
          console.log("UnSubscription successful  -> ", jsonMessage)
        }else{
          console.log("UnSubscription Failed -> ", jsonMessage)
        }
        break
      }
      case "data_response": {
        if(jsonMessage.status == "success"){
          console.log("Data Response  -> ", JSON.stringify(jsonMessage,null,2))
        }else{
          console.log("Data Request Failed -> ", jsonMessage)
        }
        break
      }
      case "stream": {
        if(jsonMessage.status == "success"){
          console.log(JSON.stringify(jsonMessage,null,2))
        }else{
          console.log("Error in stream: ", JSON.stringify(jsonMessage,null,2))
        }
        break
      }
      case "error_response": {
        if(jsonMessage.status == "success"){
          console.log(JSON.stringify(jsonMessage,null,2))
        }else{
          console.log("Error in stream: ", JSON.stringify(jsonMessage,null,2))
        }
        break
      }
    }
  }

  subscribe(){
    const subscription_request = {
      "base_token": "ETH",
      "quote_token": "USD",
      "quantity": {
        "token": "ETH",
        "levels": [0.1, 1]
      },
      "request_id": "my_request_1",
      "action": "subscribe"
    }

    this.connection.send(JSON.stringify(subscription_request));
  }

  unsubscribe(){
    const subscription_request = {
      "base_token": "ETH",
      "quote_token": "USD",
      "request_id": "my_request_2",
      "action": "unsubscribe"
    }

    this.connection.send(JSON.stringify(subscription_request));
  }

  fetchData(requestTyoe){
    const subscription_request = {
      "request_type": requestTyoe,
      "request_id": "my_request_3",
      "action": "data_request"
    }

    this.connection.send(JSON.stringify(subscription_request));
  }

  emit(event, message) {
    if (!this.connection) {
      throw "connection is not initialised yet"
    }
    this.connection.emit(event, message);
  }
}

const url = "stream.falconx.io"
const path = "/price.tickers"

apiKey = "xxx"
secretKey = "xxx"
passphrase = "xxx"

var fxStreamingClient = new FXClient(url, path, true, apiKey, passphrase, secretKey);

fxStreamingClient.connect();
