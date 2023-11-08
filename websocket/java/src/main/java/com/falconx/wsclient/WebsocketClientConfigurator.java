package com.falconx.wsclient;

import java.util.List;
import java.util.Map;

import javax.websocket.HandshakeResponse;
import javax.websocket.ClientEndpointConfig;

public class WebsocketClientConfigurator extends ClientEndpointConfig.Configurator{
    @Override
    public void beforeRequest(Map<String, List<String>> headers) {
        headers.remove("Origin");
    }

    @Override
    public void afterResponse(HandshakeResponse hr) {
        Map<String, List<String>> headers = hr.getHeaders();
        System.out.println("headers -> "+headers);
    }
}
