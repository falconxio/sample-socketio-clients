using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;
namespace SocketIOClient.Sample
{
    class Program
    {
        private const string ApiKey = "API_KEY";
        private const string SecretKey = "SECRET";
        private const string PassPhrase = "PASSPHRASE";
        private static string GetSignature(string secret, string timestamp, string method, string message)
        {
            var hasher = new HMACSHA256(Convert.FromBase64String(secret));
            hasher.Initialize();
            var hash = hasher.ComputeHash(Encoding.Default.GetBytes($"{timestamp}{method}{message}"));
            return Convert.ToBase64String(hash);
        }
        static async Task Main(string[] args)
        {
            // Show Debug and Trace messages
            //Console.OutputEncoding = Encoding.UTF8;
            //Trace.Listeners.Add(new TextWriterTraceListener(Console.Out));
            var timestamp = DateTimeOffset.Now.ToUnixTimeSeconds().ToString();
            var uri = new Uri("https://ws.falconx.io/streaming");
            var socket = new SocketIO(uri, new SocketIOOptions
            {
                Transport = Transport.TransportProtocol.WebSocket,
                AutoUpgrade = false,
                EIO = 3
            });
            socket.Options.ExtraHeaders = new Dictionary<string, string>();
            socket.Options.ExtraHeaders.Add("FX-ACCESS-SIGN", GetSignature(
                                                                SecretKey,
                                                                timestamp,
                                                                "GET",
                                                                "/socket.io/"));
            socket.Options.ExtraHeaders.Add("FX-ACCESS-TIMESTAMP", timestamp);
            socket.Options.ExtraHeaders.Add("FX-ACCESS-KEY", ApiKey);
            socket.Options.ExtraHeaders.Add("FX-ACCESS-PASSPHRASE", PassPhrase);
            socket.Options.ExtraHeaders.Add("Content-Type", "application/json");
            socket.OnConnected += Socket_OnConnected;
            socket.OnPing += Socket_OnPing;
            socket.OnPong += Socket_OnPong;
            socket.OnDisconnected += Socket_OnDisconnected;
            socket.OnReconnectAttempt += Socket_OnReconnecting;
            socket.OnError += Socket_OnError;
            socket.On("response", response =>
            {
                Console.WriteLine("Response: " + response);
            });
            socket.On("stream", response =>
            {
                Console.WriteLine("Stream: " + response);
            });
            await socket.ConnectAsync();
            Console.WriteLine("Requesting GET_ALLOWED_MARKETS configuration");
            var configRequest = new FalconXConfigRequest
            {
                message_type = "GET_ALLOWED_MARKETS",
                client_request_id = "5c5325e3-ee42-76fa-932c-64dce446d8be"
            };
            await socket.EmitAsync("request", configRequest);
            Console.WriteLine("Requesting stream of BTC-USD, quantity of 1.0");
            var pair = new Pair
            {
                base_token = "BTC",
                quote_token = "USD"
            };
            var streamRequest = new FalconXStreamRequest
            {
                token_pair = pair,
                quantity = new List<decimal> { (decimal)1.0 },
                client_request_id = "5c5325e3-ee42-76fa-932c-64dce446d8be"
            };
            await socket.EmitAsync("subscribe", streamRequest);
            Console.ReadLine();
        }
        private static void Socket_OnReconnecting(object sender, int e)
        {
            Console.WriteLine($"{DateTime.Now} Reconnecting: attempt = {e}");
        }
        private static void Socket_OnDisconnected(object sender, string e)
        {
            Console.WriteLine("disconnect: " + e);
        }
        private static async void Socket_OnConnected(object sender, EventArgs e)
        {
            Console.WriteLine("Socket_OnConnected");
            var socket = sender as SocketIO;
            Console.WriteLine("Socket.Id:" + socket.Id);
        }
        private static void Socket_OnPing(object sender, EventArgs e)
        {
            Console.WriteLine("Ping");
        }
        private static void Socket_OnPong(object sender, TimeSpan e)
        {
            Console.WriteLine("Pong: " + e.TotalMilliseconds);
        }
        private static void Socket_OnError(object sender, string error)
        {
            Console.WriteLine("Error: " + error);
        }
    }
    public class Pair
    {
        public string base_token { get; set; }
        public string quote_token { get; set; }
    }
    public class FalconXConfigRequest
    {
        public string message_type { get; set; }
        public string client_request_id { get; set; }
    }
    public class FalconXStreamRequest
    {
        public Pair token_pair { get; set; }
        public List<decimal> quantity { get; set; }
        public string client_request_id { get; set; }
    }
}
