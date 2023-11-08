using System.Net.WebSockets;
using System.Text;
using System.Security.Cryptography;
using Newtonsoft.Json;
using Newtonsoft.Json.Converters;


namespace FXWSClient.Sample
{

  public class Quantity
  {
    public string? token { get; set; }
    public double[]? levels { get; set; }
  }

  public class SubscribeRequest
  {
    public string action { get; set; } = "subscribe";
    public string? base_token { get; set; }
    public string? quote_token { get; set; }
    public Quantity? quantity { get; set; }
    public string? request_id { get; set; }
  }

  public class UnSubscribeRequest
  {
    public string action { get; set; } = "subscribe";
    public string? base_token { get; set; }
    public string? quote_token { get; set; }
    public string? request_id { get; set; }
  }

  public class DataRequest
  {
    public string action { get; set; } = "data_request";
    public string? request_type { get; set; }
    public string? request_id { get; set; }
  }

  public class FxWSResponse
  {

    public string? @event { get; set; }
    public string? status { get; set; }
    public string? request_id { get; set; }
    public object? body { get; set; }

  }

  class FalconXWSClient
  {
    private string ApiKey;
    private string SecretKey;
    private string PassPhrase;
    private string Host;
    private string Path;

    private ClientWebSocket? conn;

    public FalconXWSClient(string host, string path, string apiKey, string secret, string passphrase, bool ssl = false)
    {
      if (ssl)
      {
        this.Host = "wss://" + host;
      }
      else
      {
        this.Host = "ws://" + host;
      }
      this.Path = path;
      this.ApiKey = apiKey;
      this.SecretKey = secret;
      this.PassPhrase = passphrase;
    }

    private static string GetSignature(string secret, string timestamp, string method, string message)
    {
      var hasher = new HMACSHA256(Convert.FromBase64String(secret));
      hasher.Initialize();
      var hash = hasher.ComputeHash(Encoding.Default.GetBytes($"{timestamp}{method}{message}"));
      return Convert.ToBase64String(hash);
    }

    public async void Connect()
    {
      this.conn = new ClientWebSocket();
      Console.WriteLine("Trying to connect to URL -> " + this.Host + this.Path);
      await this.conn.ConnectAsync(new Uri(this.Host + this.Path), CancellationToken.None);
      Console.WriteLine("Connected to websocket");
      this.StartReading();
    }

    public async void Authenticate()
    {
      while (this.conn?.State != WebSocketState.Open) { }
      Dictionary<string, object> dict = new Dictionary<string, object>();

      var timestamp = DateTimeOffset.Now.ToUnixTimeSeconds();
      dict.Add("signature", GetSignature(this.SecretKey, timestamp.ToString(), "GET", this.Path));
      dict.Add("api_key", this.ApiKey);
      dict.Add("passphrase", this.PassPhrase);
      dict.Add("action", "auth");
      dict.Add("request_id", "my_sample_request");
      dict.Add("timestamp", timestamp);

      var encoded = Encoding.UTF8.GetBytes(JsonConvert.SerializeObject(dict));
      var buffer = new ArraySegment<Byte>(encoded, 0, encoded.Length);
      await this.conn.SendAsync(buffer, WebSocketMessageType.Text, true, CancellationToken.None);
    }

    public async void Subscribe()
    {
      if (this.conn?.State == WebSocketState.Open)
      {

        SubscribeRequest subscriberequest = new SubscribeRequest
        {
          action = "subscribe",
          base_token = "ETH",
          quote_token = "USD",
          quantity = new Quantity
          {
            token = "ETH",
            levels = new double[] { 0.1, 1, 3, 4 }
          },
          request_id = "my_sample_request_1"
        };

        Console.WriteLine(JsonConvert.SerializeObject(subscriberequest));
        var encoded = Encoding.UTF8.GetBytes(JsonConvert.SerializeObject(subscriberequest));
        var buffer = new ArraySegment<Byte>(encoded, 0, encoded.Length);
        await this.conn.SendAsync(buffer, WebSocketMessageType.Text, true, CancellationToken.None);

      }
      else
      {
        Console.WriteLine("Not connected.");
      }
    }

    public async void UnSubscribe()
    {
      if (this.conn?.State == WebSocketState.Open)
      {

        UnSubscribeRequest unSubscriberequest = new UnSubscribeRequest
        {
          action = "unsubscribe",
          base_token = "ETH",
          quote_token = "USD",
          request_id = "my_sample_request_1"
        };

        Console.WriteLine(JsonConvert.SerializeObject(unSubscriberequest));
        var encoded = Encoding.UTF8.GetBytes(JsonConvert.SerializeObject(unSubscriberequest));
        var buffer = new ArraySegment<Byte>(encoded, 0, encoded.Length);
        await this.conn.SendAsync(buffer, WebSocketMessageType.Text, true, CancellationToken.None);

      }
      else
      {
        Console.WriteLine("Not connected.");
      }
    }

    public async void FetchData(string RequestType)
    {
      if (this.conn?.State == WebSocketState.Open)
      {

        DataRequest unSubscriberequest = new DataRequest
        {
          action = "data_request",
          request_type = RequestType,
          request_id = "my_sample_request_1"
        };

        Console.WriteLine(JsonConvert.SerializeObject(unSubscriberequest));
        var encoded = Encoding.UTF8.GetBytes(JsonConvert.SerializeObject(unSubscriberequest));
        var buffer = new ArraySegment<Byte>(encoded, 0, encoded.Length);
        await this.conn.SendAsync(buffer, WebSocketMessageType.Text, true, CancellationToken.None);
      }
      else
      {
        Console.WriteLine("Not connected.");
      }
    }

    public async void StartReading()
    {
      byte[] buf = new byte[1056];
      while (this.conn?.State == WebSocketState.Open || this.conn?.State == WebSocketState.CloseSent)
      {
        var result = await this.conn.ReceiveAsync(buf, CancellationToken.None);
        if (result.MessageType == WebSocketMessageType.Close)
        {
          await this.conn.CloseAsync(WebSocketCloseStatus.NormalClosure, null, CancellationToken.None);
          Console.WriteLine(result.CloseStatusDescription);
        }
        else
        {
          string jsonString = Encoding.ASCII.GetString(buf, 0, result.Count);
          if (jsonString != null)
          {
            FxWSResponse ex = JsonConvert.DeserializeObject<FxWSResponse>(jsonString);

            switch (ex?.@event)
            {
              case "auth_response":
                {
                  if (ex?.status == "success")
                  {
                    Console.WriteLine("Authentication Successfull", ex);
                  }
                  else
                  {
                    Console.WriteLine("Authentication Failed", ex);
                  }
                  break;
                }
              case "subscribe_response":
                {
                  if (ex?.status == "success")
                  {
                    Console.WriteLine("Subscription Successfull", ex);
                  }
                  else
                  {
                    Console.WriteLine("Subscription Failed", ex);
                  }
                  break;
                }
              case "unsubscribe_response":
                {
                  if (ex?.status == "success")
                  {
                    Console.WriteLine("UnSubscription Successfull", ex);
                  }
                  else
                  {
                    Console.WriteLine("UnSubscription Failed", ex);
                  }
                  break;
                }
              case "data_response":
                {
                  if (ex?.status == "success")
                  {
                    Console.WriteLine("Data Response");
                    var prettyJson = JsonConvert.SerializeObject(
                      ex.body, Formatting.Indented,
                      new JsonConverter[] { new StringEnumConverter() }
                    );

                    Console.WriteLine(prettyJson);
                  }
                  else
                  {
                    Console.WriteLine("Data Request Failed", ex);
                  }
                  break;
                }
              case "error_response":
                {
                  Console.WriteLine("Error Received", ex);
                  break;
                }
              case "stream":
                {
                  if (ex?.status == "success")
                  {
                    var prettyJson = JsonConvert.SerializeObject(
               ex.body, Formatting.Indented,
               new JsonConverter[] { new StringEnumConverter() });

                    Console.WriteLine(prettyJson);
                  }
                  break;
                }
              default:
                {
                  Console.WriteLine("Unknown response type", ex);
                  break;
                }
            }
          }
        }
      }
    }
  }

  class Program
  {
    static void Main(string[] args)
    {
      string apiKey = "xxx";
      string secretKey = "xxx";
      string passPhrase = "xxx";
      string host = "stream.falconx.io";
      string path = "/price.tickers";

      var exitEvent = new ManualResetEvent(false);

      var fx_client = new FalconXWSClient(host, path, apiKey, secretKey, passPhrase, true);

      fx_client.Connect();

      fx_client.Authenticate();

      fx_client.Subscribe();

      // fx_client.FetchData("max_levels");
      // fx_client.FetchData("max_connections");
      exitEvent.WaitOne();
    }
  }

}