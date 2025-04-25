package com.careerup.llmgateway.client;

import com.careerup.llmgateway.config.DeepseekConfig;
import com.careerup.llmgateway.exception.DeepseekException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.util.Map;

@Slf4j
@Component
@RequiredArgsConstructor
public class DeepseekClient {
    private final RestTemplate restTemplate;
    private final DeepseekConfig config;

    public String generateCompletion(String prompt) {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.setBearerAuth(config.getApiKey());

        Map<String, Object> requestBody = Map.of(
            "model", config.getModel(),
            "prompt", prompt,
            "max_tokens", 1000,
            "temperature", 0.7
        );

        HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestBody, headers);

        try {
            String response = restTemplate.postForObject(
                config.getBaseUrl() + "/completions",
                request,
                String.class
            );
            log.info("Successfully generated completion from Deepseek");
            return response;
        } catch (Exception e) {
            log.error("Error generating completion from Deepseek: {}", e.getMessage());
            throw new DeepseekException("Failed to generate completion", e);
        }
    }
} 