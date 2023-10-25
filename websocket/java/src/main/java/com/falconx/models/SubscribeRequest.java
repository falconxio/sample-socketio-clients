package com.falconx.models;

import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

public class SubscribeRequest {
  public String action;

    @JsonProperty("base_token")
    public String baseToken;

    @JsonProperty("quote_token")
    public String quoteToken;

    public Quantity quantity;

    @JsonProperty("request_id")
    public String requestId;

    public SubscribeRequest(String action, String baseToken, String quoteToken, String quantityToken,
        List<Double> levels, String requestId) {
      this.action = action;
      this.baseToken = baseToken;
      this.quoteToken = quoteToken;
      this.quantity = new Quantity(quantityToken, levels);
      this.requestId = requestId;
    }
}
