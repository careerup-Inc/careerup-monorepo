package com.careerup.llmgateway.util;

import org.springframework.stereotype.Component;

@Component
public class TokenCounter {
    // Simple approximation: 1 token ≈ 4 characters
    private static final int CHARS_PER_TOKEN = 4;

    public long countTokens(String text) {
        if (text == null || text.isEmpty()) {
            return 0;
        }
        return (text.length() + CHARS_PER_TOKEN - 1) / CHARS_PER_TOKEN;
    }
} 