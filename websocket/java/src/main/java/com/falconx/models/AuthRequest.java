package com.falconx.models;

import com.fasterxml.jackson.annotation.JsonProperty;

public class AuthRequest {
  public String action;

    @JsonProperty("api_key")
    public String apiKey;

    public String passphrase;

    public String signature;

    public Long timestamp;

    @JsonProperty("request_id")
    public String requestId;

    public AuthRequest(String apiKey, String passphrase, String sign, Long timestamp,
        String requestId) {
      this.action = "auth";
      this.apiKey = apiKey;
      this.passphrase = passphrase;
      this.signature = sign;
      this.timestamp = timestamp;
      this.requestId = requestId;
    }
}
