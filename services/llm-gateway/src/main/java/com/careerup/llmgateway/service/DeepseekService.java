package com.careerup.llmgateway.service;

import com.careerup.llmgateway.client.DeepseekClient;
import com.careerup.llmgateway.model.CompletionRequest;
import com.careerup.llmgateway.model.CompletionResponse;
import com.careerup.llmgateway.util.TokenCounter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

@Slf4j
@Service
@RequiredArgsConstructor
public class DeepseekService {
    private final DeepseekClient deepseekClient;
    private final TokenCounter tokenCounter;

    public CompletionResponse processPrompt(CompletionRequest request) {
        log.info("Processing prompt with Deepseek: {}", request.getPrompt());
        
        long promptTokens = tokenCounter.countTokens(request.getPrompt());
        String completion = deepseekClient.generateCompletion(request.getPrompt());
        long completionTokens = tokenCounter.countTokens(completion);
        
        return CompletionResponse.builder()
                .completion(completion)
                .model("deepseek-r1")
                .tokensUsed(promptTokens + completionTokens)
                .build();
    }
} 