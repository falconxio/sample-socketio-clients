import argparse
import base64
import hashlib
import hmac
import ssl
import time
import uuid
import logzero
import socketio

logger = logzero.logger

ctx = ssl.create_default_context()
ctx.set_ciphers('DEFAULT')

URL = 'https://stream.falconx.io'
API_KEY = '<Your API_KEY>'
PASSPHRASE = '<Your PASSPHRASE>'
SECRET_KEY = '<Your SECRET_KEY>'

def create_header(api_key, secret_key, passphrase):
    timestamp = str(time.time())
    message = timestamp + 'GET' + "/socket.io/"
    hmac_key = base64.b64decode(secret_key)
    signature = hmac.new(hmac_key, message.encode(), hashlib.sha256)
    signature_b64 = base64.b64encode(signature.digest())
    return {
        'FX-ACCESS-SIGN': signature_b64.decode("utf-8"),
        'FX-ACCESS-TIMESTAMP': timestamp,
        'FX-ACCESS-KEY': api_key,
        'FX-ACCESS-PASSPHRASE': passphrase,
        'Content-Type': 'application/json',
    }


class FastRFSClient(socketio.ClientNamespace):
    def __init__(self, namespace):
        self.subscription_requests = []
        self.previous_response_times = {}
        self.running_avg_times = {}
        super().__init__(namespace)

    def populate_subscription_requests(self, token_pairs: list, levels: list):
        for token_pair in token_pairs:
            base_token, quote_token = token_pair.split("/")
            self.subscription_requests.append({
                'token_pair': {
                    'base_token': base_token,
                    'quote_token': quote_token
                },
                'quantity': levels,
                'quantity_token': quote_token,  # only for v2/subscribe (optional)
                'client_request_id': str(uuid.uuid4()),
                'echo_id': True
            })

    def on_connect(self):
        logger.info('Server connected.')
        for subscription_request in self.subscription_requests:
            self.emit('subscribe', subscription_request, namespace='/streaming')
            client_request_id = subscription_request['client_request_id']
            self.previous_response_times[client_request_id] = time.time()
        logger.info('Finished subscribing.')

    def on_disconnect(self, *args):
        logger.info('Server disconnected.' + str(args))

    def on_connect_error(self, *args):
        logger.info("Cannot connect to the server. Error", args)

    def on_response(self, *args):
        response_time = time.time()
        client_request_id = args[0]['client_request_id']
        logger.info("Client received response: " + str(args))
        previous_response_time = self.previous_response_times[client_request_id]
        logger.info(f'Client_request_id: {client_request_id}, time taken for first response {response_time - previous_response_time}')
        self.previous_response_times[client_request_id] = response_time

    def on_stream(self, *args):
        response_time = time.time()
        client_request_id = args[0][0]['client_request_id']
        logger.info("printing arguments ", *args)
        self.previous_response_times[client_request_id] = response_time

    def on_error(self, *args):
        logger.info("error", args)


def main(args):
    client = FastRFSClient(namespace='/streaming')
    client.populate_subscription_requests(args.token_pairs, args.levels)
    headers = create_header(API_KEY, SECRET_KEY, PASSPHRASE)
    socketio_client = socketio.Client(logger=False, engineio_logger=False, ssl_verify=False)
    socketio_client.register_namespace(client)
    socketio_client.connect(URL, namespaces=['/streaming'], transports=['websocket'], headers=headers)

    while True:
        continue


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--token_pairs',
                        type=lambda s: s.split(","),
                        required=False,
                        default="BTC/USD",
                        help="Comma separated token pairs eg. BTC/USD,ETH/USD")
    parser.add_argument('--levels',
                        type=lambda s: [int(x) for x in s.split(",")],
                        required=False,
                        default="1,2",
                        help="Comma separated levels eg. 190,200,400")

    args = parser.parse_args()
    main(args)
