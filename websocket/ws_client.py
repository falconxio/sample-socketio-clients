import websocket
from marshmallow import fields
from marshmallow.validate import Range
import base64
import hashlib
import hmac
import time
import json
from pprint import pprint

class ConnectionOpts:
    retry_on_error = fields.Boolean(default=False)
    num_retries = fields.Integer(validate=Range(min=1, error="Value must be greater than 0"))
    retry_delay = fields.Integer()
    log_streams = fields.Boolean(default=False)

    def __init__(self,
                 retry_on_error = False,
                 num_retries = None,
                 retry_delay = None,
                 log_streams = False,
                 ) -> None:
        self.retry_on_error = retry_on_error
        self.num_retries = num_retries
        self.retry_delay = retry_delay
        self.log_streams = log_streams


class FalconXWSClient:
    host = fields.String(required=True)
    path = fields.String()
    ssl = fields.Boolean(default=False)
    api_key = fields.String(required=True)
    secret = fields.String(required=True)
    passphrase = fields.String(required=True)
    connection_opts = fields.Nested(ConnectionOpts)
    authenticated = fields.Boolean(default=False)
    reader_active = fields.Boolean(default=False)
    conn = fields.Nested(websocket.WebSocketApp)
    retry_count = 0

    def __init__(self,
                 host: str = None,
                 path: str = "/",
                 ssl:bool = False,
                 api_key: str = None,
                 secret: str = None,
                 passphrase: str = None,
                 connection_opts: ConnectionOpts = None
                 ) -> None:
        self.ssl = ssl
        self.path = path
        self.host = ("wss://" if ssl else "ws://") + host
        self.api_key = api_key
        self.secret = secret
        self.passphrase = passphrase
        self.connection_opts = connection_opts

    def get_signature(self, timestamp, path):
        method = "GET"

        message = "".join([timestamp, method, path, ""])
        hmac_key = base64.b64decode(self.secret)
        signature = hmac.new(hmac_key, message.encode(), hashlib.sha256)
        signature_b64 = base64.b64encode(signature.digest())
        signature = signature_b64.decode("utf-8")
        return signature

    def connect(self):
        def on_message(ws, message):
            # print(message)
            data = json.loads(message)
            if data["event"] == "auth_response" and data["status"] == "success":
                self.authenticated = True
                self.susbcribe()
            elif data["event"] == "stream" and data["status"] == "success":
                pprint(data["body"])

        def on_error(ws, error):
            print("Error : ", error)

        def on_close(ws, close_status_code, close_msg):
            print("### closed ###")

        def on_open(ws):
            print("Opened connection")
            self.authenticate()

        # websocket.enableTrace(True)
        self.conn = websocket.WebSocketApp(self.host,
                                on_open=on_open,
                                on_message=on_message,
                                on_error=on_error,
                                on_close=on_close)
        self.conn.run_forever(reconnect=5)
    
    def authenticate(self):
        timestamp = int(time.time())
        signature = self.get_signature(str(timestamp), self.path)

        req = {
          "action":     "auth",
          "api_key":    self.api_key,
          "passphrase": self.passphrase,
          "secret_key": self.secret,
          "sign":       signature,
          "timestamp":  timestamp,
          "request_id": "my_request"
        }
        self.conn.send(data=json.dumps(req))

    def susbcribe(self):
        susbcription_request = {
            "base_token": "ETH",
            "quote_token": "USD",
            "quantity": {
                "token": "ETH",
                "levels": [0.1, 0.1]
            },
            "request_id": "my_request_1",
            "action": "subscribe"
        }
        self.conn.send(data=json.dumps(susbcription_request))


if __name__ == "__main__":
    host = "stream.falconx.io/price.tickers"
    path = "/price.tickers"
    api_key = 'xxx'
    passphrase = "xxx"
    secret_key = "xxx"
    
    connection_opts = ConnectionOpts(retry_on_error=True, num_retries=5, retry_delay=1)

    fx_ws_client = FalconXWSClient(host=host,path=path, api_key=api_key, secret=secret_key, passphrase=passphrase, connection_opts=connection_opts)

    fx_ws_client.connect()
    
