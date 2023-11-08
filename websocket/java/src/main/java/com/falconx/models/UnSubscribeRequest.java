package com.falconx.models;

import com.fasterxml.jackson.annotation.JsonProperty;

public class UnSubscribeRequest {
  public String action;

    @JsonProperty("base_token")
    public String baseToken;

    @JsonProperty("quote_token")
    public String quoteToken;

    @JsonProperty("request_id")
    public String requestId;

    public UnSubscribeRequest(String baseToken, String quoteToken, String requestId) {
      this.action = "unsubscribe";
      this.baseToken = baseToken;
      this.quoteToken = quoteToken;
      this.requestId = requestId;
    }
}
