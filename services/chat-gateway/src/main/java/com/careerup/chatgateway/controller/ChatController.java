package com.careerup.chatgateway.controller;

import com.careerup.chatgateway.model.ChatMessage;
import com.careerup.chatgateway.service.ChatService;
import lombok.RequiredArgsConstructor;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Controller;

@Controller
@RequiredArgsConstructor
public class ChatController {

    private final SimpMessagingTemplate messagingTemplate;
    private final ChatService chatService;

    @MessageMapping("/chat.send")
    public void sendMessage(@Payload ChatMessage chatMessage) {
        // Process the message through the chat service
        ChatMessage processedMessage = chatService.processMessage(chatMessage);
        
        // Send the processed message to the appropriate topic
        messagingTemplate.convertAndSend("/topic/public", processedMessage);
    }
} 