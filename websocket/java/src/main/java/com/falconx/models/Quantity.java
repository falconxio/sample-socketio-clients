package com.falconx.models;

import java.util.List;

public class Quantity {
  public String token;
    public List<Double> levels;

    public Quantity(String token, List<Double> levels) {
      this.levels = levels;
      this.token = token;
    }
}
