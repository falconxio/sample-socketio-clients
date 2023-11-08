package com.falconx.models;

import com.fasterxml.jackson.annotation.JsonProperty;

public class DataRequest {
  public String action;

    @JsonProperty("request_type")
    public String requestType;

    @JsonProperty("request_id")
    public String requestId;

    public DataRequest(String requestType, String requestId) {
      this.action = "data_request";
      this.requestType = requestType;
      this.requestId = requestId;
    }
}
