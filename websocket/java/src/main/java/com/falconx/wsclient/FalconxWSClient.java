package com.falconx.wsclient;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.concurrent.CountDownLatch;

import javax.crypto.spec.SecretKeySpec;
import javax.websocket.ClientEndpoint;
import javax.websocket.CloseReason;
import javax.websocket.DeploymentException;
import javax.websocket.OnClose;
import javax.websocket.OnMessage;
import javax.websocket.OnOpen;
import javax.websocket.Session;

import org.glassfish.tyrus.client.ClientManager;

import java.util.ArrayList;
import java.util.Base64;
import java.util.Date;
import javax.crypto.Mac;

import com.falconx.models.AuthRequest;
import com.falconx.models.Response;
import com.falconx.models.SubscribeRequest;
import com.falconx.models.UnSubscribeRequest;
import com.fasterxml.jackson.annotation.*;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;



@ClientEndpoint(configurator = WebsocketClientConfigurator.class)
public class FalconxWSClient {

  private static CountDownLatch latch;

  // public void onConnect;
  static String host = "stream.falconx.io";
  static String path = "/price.tickers";
  public String apiKey;
  public String secretKey;
  public String passphrase;

  public static String GetConnectURL(boolean ssl){
    return (ssl ? "wss://":"ws://" ) + FalconxWSClient.host + FalconxWSClient.path;
  }

  private String generateSignature() throws Exception {
    long timestamp = new Date().getTime() / 1000;

    String preHash = timestamp + "GET" + FalconxWSClient.path;
    SecretKeySpec keyspec = new SecretKeySpec(Base64.getDecoder().decode(secretKey), "HmacSHA256");
    Mac sha256 = Mac.getInstance("HmacSHA256");
    sha256.init(keyspec);
    return Base64.getEncoder().encodeToString(sha256.doFinal(preHash.getBytes()));
  }

  public void Authenticate(Session session) {
    long timestamp = new Date().getTime() / 1000;

    try {
      AuthRequest authRequest = new AuthRequest(
          "auth",
          this.apiKey,
          this.passphrase,
          this.generateSignature(),
          timestamp,
          "my_request_id_1");

      ObjectMapper mapper = new ObjectMapper();
      System.out.println(mapper.writeValueAsString(authRequest));
      session.getBasicRemote().sendText(mapper.writeValueAsString(authRequest));
    } catch (Exception ex) {
      System.out.println(ex);
    }
  }

  public String Subscribe(Session session) {
    List<Double> levels = new ArrayList<>();
    levels.add(0.5);
    levels.add(1.0);
    levels.add(5.0);
    SubscribeRequest request = new SubscribeRequest("subscribe", "ETH", "USD", "ETH", levels, "my_subscribe_request");
    ObjectMapper mapper = new ObjectMapper();
    try {
      return mapper.writeValueAsString(request);
    } catch (JsonProcessingException e) {
      e.printStackTrace();
    }
    return "";
  }

  public String UnSubscribe(Session session) {
    UnSubscribeRequest request = new UnSubscribeRequest("unsubscribe", "ETH", "USD", "my_unsubscribe_request");
    ObjectMapper mapper = new ObjectMapper();
    try {
      return mapper.writeValueAsString(request);
    } catch (JsonProcessingException e) {
      e.printStackTrace();
    }
    return "";
  }

  @OnOpen
  public void onOpen(Session session) {
    System.out.println("--- Connected " + session.getId());
    this.SetupConfig();
    this.Authenticate(session);
  }

  @OnMessage
  public String onMessage(String message, Session session) {
    try {
      System.out.println("--- Received " + message);

      ObjectMapper mapper = new ObjectMapper();
      Response response = mapper.readValue(message, Response.class);
      System.out.println("Data " + response.GetEvent());
      switch (response.GetEvent()) {
        case "auth_response": {
          if (response.GetStatus().compareToIgnoreCase("success") == 0) {
            return this.Subscribe(session);
          } else {
            System.out.println("Authentication Failed: " + message);
          }
          break;
        }
        case "subscribe_response": {
          if (response.GetStatus().compareToIgnoreCase("success") == 0) {
            System.out.println(message);
          } else {
            System.out.println("Authentication Failed: " + message);
          }

          break;
        }
        case "unsubscribe_response": {
          if (response.GetStatus().compareToIgnoreCase("success") == 0) {
            System.out.println(message);
          } else {
            System.out.println("Authentication Failed: " + message);
          }
          break;
        }
        case "stream": {
          System.out.println(message);
          break;
        }
      }
      return null;
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @OnClose
  public void onClose(Session session, CloseReason closeReason) {
    System.out.println("Session " + session.getId() +
        " closed because " + closeReason);
    latch.countDown();
  }

  public void SetupConfig() {
    this.apiKey = "xxx";
    this.secretKey = "xxx";
    this.passphrase = "xxx";
  }

  public static void main(String[] args) {
    latch = new CountDownLatch(1);
    ClientManager client = ClientManager.createClient();

    try {
      URI uri = new URI(FalconxWSClient.GetConnectURL(true));
      client.connectToServer(FalconxWSClient.class, uri);
      latch.await();
    } catch (DeploymentException | URISyntaxException | InterruptedException e) {
      e.printStackTrace();
    } catch (IOException e) {
      e.printStackTrace();
    }
  }
}