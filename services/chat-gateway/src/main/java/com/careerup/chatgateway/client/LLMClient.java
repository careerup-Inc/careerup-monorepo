package com.careerup.chatgateway.client;

import com.careerup.chatgateway.model.ChatMessage;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

@Slf4j
@Component
@RequiredArgsConstructor
public class LLMClient {
    private final RestTemplate restTemplate;
    
    @Value("${llm.service.addr}")
    private String llmServiceAddr;

    public String processMessage(ChatMessage message) {
        String url = "http://" + llmServiceAddr + "/api/v1/deepseek/completion";
        
        try {
            var response = restTemplate.postForObject(
                url,
                message.getContent(),
                String.class
            );
            log.info("Successfully processed message with LLM service");
            return response;
        } catch (Exception e) {
            log.error("Error processing message with LLM service: {}", e.getMessage());
            return "I apologize, but I'm having trouble processing your message right now.";
        }
    }
} 