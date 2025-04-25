package com.careerup.chatgateway.service;

import com.careerup.chatgateway.model.ChatMessage;

public interface ChatService {
    ChatMessage processMessage(ChatMessage message);
    void sendMessage(ChatMessage message);
} 