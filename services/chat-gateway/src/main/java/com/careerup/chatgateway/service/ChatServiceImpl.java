package com.careerup.chatgateway.service;

import com.careerup.chatgateway.client.LLMClient;
import com.careerup.chatgateway.model.ChatMessage;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;

@Slf4j
@Service
@RequiredArgsConstructor
public class ChatServiceImpl implements ChatService {
    private final SimpMessagingTemplate messagingTemplate;
    private final LLMClient llmClient;

    @Override
    public ChatMessage processMessage(ChatMessage message) {
        log.info("Processing message: {}", message);
        
        // Get response from LLM service
        String llmResponse = llmClient.processMessage(message);
        
        // Create response message
        ChatMessage response = ChatMessage.builder()
                .sender("AI")
                .content(llmResponse)
                .type(ChatMessage.MessageType.CHAT)
                .build();
        
        sendMessage(response);
        return response;
    }

    @Override
    public void sendMessage(ChatMessage message) {
        log.info("Sending message: {}", message);
        messagingTemplate.convertAndSend("/topic/public", message);
    }
} 