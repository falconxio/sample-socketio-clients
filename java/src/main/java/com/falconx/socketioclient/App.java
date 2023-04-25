package com.falconx.socketioclient;

import java.util.HashMap;
import java.util.Map;
import javax.crypto.spec.SecretKeySpec;
import java.util.Base64;
import java.util.Date;
import javax.crypto.Mac;

/**
 * Main App
 */
public final class App {
    private App() {
    }

    /**
     * Sample Java client for streaming prices
     * 
     * @param args The arguments of the program.
     */
    public static void main(String[] args) throws Exception {
        Map<String, String> headers = getKeys("SECRET",
                "API_KEY", "PASSPHRASE");

        SocketIOClient client = new SocketIOClient("https://ws-stream.falconx.io", headers);
        client.connect();

    }

    private static Map<String, String> getKeys(String secretKey, String apiKey, String passphrase) throws Exception {
        long timestamp = new Date().getTime() / 1000;

        String preHash = timestamp + "GET" + "/socket.io/";
        SecretKeySpec keyspec = new SecretKeySpec(Base64.getDecoder().decode(secretKey), "HmacSHA256");
        Mac sha256 = Mac.getInstance("HmacSHA256");
        sha256.init(keyspec);
        String signature = Base64.getEncoder().encodeToString(sha256.doFinal(preHash.getBytes()));

        Map<String, String> headers = new HashMap<>();
        headers.put("FX-ACCESS-SIGN", signature);
        headers.put("FX-ACCESS-TIMESTAMP", Long.toString(timestamp));
        headers.put("FX-ACCESS-KEY", apiKey);
        headers.put("FX-ACCESS-PASSPHRASE", passphrase);
        headers.put("Content-Type", "application/json");
        return headers;
    }
}
