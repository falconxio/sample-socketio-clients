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
      'request_id': "my_auth_request_1",
      'sign': signatureB64,
      'timestamp': timestamp,
      'api_key': this.apiKey,
      'passphrase': this.passphrase,
    }
  }

  heartbeat(client) {
    clearTimeout(client.pingTimeout);

    // Use `WebSocket#terminate()`, which immediately destroys the connection,
    // instead of `WebSocket#close()`, which waits for the close timer.
    // Delay should be equal to the interval at which your server
    // sends out pings plus a conservative assumption of the latency.
    client.pingTimeout = setTimeout(() => {
      client.terminate();
    }, 30000 + 1000);
  }

  onConnect() {
    console.log(`Connected to websocket server`);
    this.authenticate()
  }

  authenticate() {
    const authRequest = this.createAuthRequest()
    this.connection.send(JSON.stringify(authRequest))
  }

  onDisconnect() {
    console.log(`Disconnected to ${this.namespace} namespace`);
    console.log('Retrying connection..')
    this.connect();
  }

  onError(msg) {
    console.log(`Error. Message received:`, msg)
  }

  connect() {
    if (this.connection) {
    }
    const protocol = this.ssl ? "wss" : "ws"
    const finalUrl = `${protocol}://${this.url}${this.path}`
    console.log(finalUrl)
    this.connection = new WebSocket(finalUrl);
    this.connection.on('open', this.onConnect.bind(this));
    this.connection.on('error', this.onError.bind(this));
    this.connection.on('ping', this.heartbeat.bind(this));
    this.connection.on('message', this.onMessage.bind(this));
  }

  onMessage(msg) {
    const stringData = String.fromCharCode.apply(null, msg)
    const jsonMessage = JSON.parse(stringData)
    switch(jsonMessage.event) {
      case "auth_response": {
        if(jsonMessage.status == "success"){
          this.subscribe()
        }else{
          console.log("Authentication Failure")
        }
      }
      case "subscribe_response": {
        if(jsonMessage.status == "success"){
          console.log("Subscription successful  -> ", jsonMessage)
        }else{
          console.log("Subscription Failed -> ", jsonMessage)
        }
      }
      case "unsubscribe_response": {
        if(jsonMessage.status == "success"){
          console.log("UnSubscription successful  -> ", jsonMessage)
        }else{
          console.log("UnSubscription Failed -> ", jsonMessage)
        }
      }
      case "data_response": {
        if(jsonMessage.status == "success"){
          console.log("Data Response  -> ", jsonMessage)
        }else{
          console.log("Data Request Failed -> ", jsonMessage)
        }
      }
      case "stream": {
        if(jsonMessage.status == "success"){
          console.log(JSON.stringify(jsonMessage,null,2))
        }else{
          console.log("Error in stream: ", JSON.stringify(jsonMessage,null,2))
        }
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

var fxStreamingClient = new FXClient(url, path, false, apiKey, passphrase, secretKey);

fxStreamingClient.connect();
