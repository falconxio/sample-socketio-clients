package com.falconx.socketioclient;

import io.socket.client.IO;
import io.socket.client.Manager;
import io.socket.client.Socket;
import io.socket.emitter.Emitter;
import io.socket.engineio.client.transports.Polling;
import io.socket.engineio.client.transports.WebSocket;
import okhttp3.OkHttpClient;
import okhttp3.Request;


import java.net.URI;
import java.net.URISyntaxException;
import java.util.Map;

import org.json.JSONArray;
import org.json.JSONObject;

public class SocketIOClient {

    public Socket socket;

    public SocketIOClient(String url, Map<String, String> headers) throws URISyntaxException {
        OkHttpClient.Builder builder = new OkHttpClient.Builder();
        // Add custom headers to the HTTP request
        if (headers != null) {
            builder.addInterceptor(chain -> {
                Request request = chain.request();
                for (Map.Entry<String, String> entry : headers.entrySet()) {
                    request = request.newBuilder()
                            .addHeader(entry.getKey(), entry.getValue())
                            .build();
                }
                return chain.proceed(request);
            });
        }
        OkHttpClient client = builder.build();
        IO.Options options = new IO.Options();
        options.transports = new String[] { WebSocket.NAME, Polling.NAME };
        options.callFactory = client;
        options.webSocketFactory = client;
        System.out.println(url);
        Manager manager = new Manager(URI.create(url), options);
        this.socket = manager.socket("/streaming", options);

        // Register event listeners
        Socket socket = this.socket;
        this.socket.on(Socket.EVENT_CONNECT, new Emitter.Listener() {
            @Override
            public void call(Object... args) {
                System.out.println("Connected to FalconX server");
                try {
                    // Sample Subscription Request
                    JSONObject sub_req = new JSONObject();
                    sub_req.put("quantity_token", "BTC");
                    sub_req.put("client_request_id", "client_request_id");

                    JSONArray quantity = new JSONArray();
                    quantity.put(1.0);
                    quantity.put(5.0);
                    JSONObject token_pair = new JSONObject();
                    token_pair.put("base_token", "BTC");
                    token_pair.put("quote_token", "USD");

                    sub_req.put("token_pair", token_pair);
                    sub_req.put("quantity", quantity);
                    socket.emit("subscribe", sub_req);
                } catch (Exception e) {
                    System.out.println(e);
                }
            }
        }).on(Socket.EVENT_DISCONNECT, new Emitter.Listener() {
            @Override
            public void call(Object... args) {
                System.out.println("Disconnected from FalconX server");
            }
        }).on("stream", new Emitter.Listener() {
            @Override
            public void call(Object... args) {
                System.out.println("stream");
                System.out.println(System.currentTimeMillis());
                JSONArray price = null;
                for (Object arg : args) {
                    if (arg instanceof JSONArray) {
                        price = (JSONArray) arg;
                        break;
                    }
                }
                System.out.println(price);
            }
        }).on(Socket.EVENT_CONNECT_ERROR, new Emitter.Listener() {
            @Override
            public void call(Object... args) {
                System.out.println(" EVENT_CONNECT_ERROR connect error");

            }
        });

    }

    public void connect() throws Exception {
        this.socket.connect();
        System.out.println("connect req sent");
    }

    public void sendMessage(String event, Object... args) {
        this.socket.emit(event, args);
    }

    public void disconnect() {
        this.socket.disconnect();
    }
}
