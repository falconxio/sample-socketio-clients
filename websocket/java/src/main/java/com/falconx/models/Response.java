package com.falconx.models;

import com.fasterxml.jackson.annotation.JsonProperty;

public class Response {
  @JsonProperty("status")
    String status;
    @JsonProperty("event")
    String event;
    @JsonProperty("request_id")
    String requestId;
    @JsonProperty("body")
    Object body;
    @JsonProperty("error")
    Object error;

    public Response() {
    }

    public String GetEvent(){
      return this.event;
    }

    public Object GetBody(){
      return this.body;
    }
    public String GetStatus(){
      return this.status;
    }
}
