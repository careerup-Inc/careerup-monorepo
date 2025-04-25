package com.careerup.authcore.model.dto;

public record TokenResponse(String accessToken, String refreshToken, long expiresIn) {}
