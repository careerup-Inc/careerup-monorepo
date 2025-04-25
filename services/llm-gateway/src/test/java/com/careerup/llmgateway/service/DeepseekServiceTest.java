package com.careerup.llmgateway.service;

import com.careerup.llmgateway.client.DeepseekClient;
import com.careerup.llmgateway.model.CompletionRequest;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class DeepseekServiceTest {
    @Mock
    private DeepseekClient deepseekClient;
    
    @InjectMocks
    private DeepseekService deepseekService;

    @Test
    void processPrompt_Success() {
        // TODO finish this test
        String expectedResponse = "Test completion";
        CompletionRequest request = new CompletionRequest();
        request.setPrompt("test prompt");
        
        when(deepseekClient.generateCompletion(anyString())).thenReturn(expectedResponse);
        
        String result = deepseekService.processPrompt(request).toString();
        
        assertEquals(expectedResponse, result);
    }
}