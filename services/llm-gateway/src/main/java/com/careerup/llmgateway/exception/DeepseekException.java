package com.careerup.llmgateway.exception;

public class DeepseekException extends RuntimeException {
    public DeepseekException(String message) {
        super(message);
    }

    public DeepseekException(String message, Throwable cause) {
        super(message, cause);
    }
} 