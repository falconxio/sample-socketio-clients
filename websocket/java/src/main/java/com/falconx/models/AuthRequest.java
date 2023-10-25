package com.falconx.models;

import com.fasterxml.jackson.annotation.JsonProperty;

public class AuthRequest {
  public String action;

    @JsonProperty("api_key")
    public String apiKey;

    public String passphrase;

    public String sign;

    public Long timestamp;

    @JsonProperty("request_id")
    public String requestId;

    public AuthRequest(String action, String apiKey, String passphrase, String sign, Long timestamp,
        String requestId) {
      this.action = action;
      this.apiKey = apiKey;
      this.passphrase = passphrase;
      this.sign = sign;
      this.timestamp = timestamp;
      this.requestId = requestId;
    }
}
