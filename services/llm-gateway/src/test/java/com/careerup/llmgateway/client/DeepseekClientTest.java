package com.careerup.llmgateway.client;

import com.careerup.llmgateway.config.DeepseekConfig;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.http.HttpMethod;
import org.springframework.http.MediaType;
import org.springframework.test.web.client.MockRestServiceServer;
import org.springframework.web.client.RestTemplate;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.springframework.test.web.client.match.MockRestRequestMatchers.*;
import static org.springframework.test.web.client.response.MockRestResponseCreators.withSuccess;

class DeepseekClientTest {
    private DeepseekClient deepseekClient;
    private MockRestServiceServer mockServer;
    private RestTemplate restTemplate;

    @BeforeEach
    void setUp() {
        restTemplate = new RestTemplate();
        mockServer = MockRestServiceServer.createServer(restTemplate);
        
        DeepseekConfig config = new DeepseekConfig();
        config.setApiKey("test-api-key");
        config.setBaseUrl("http://test-api");
        config.setModel("deepseek-r1");
        
        deepseekClient = new DeepseekClient(restTemplate, config);
    }

    @Test
    void generateCompletion_Success() {
        String expectedResponse = "{\"choices\":[{\"text\":\"Test completion\"}]}";
        mockServer.expect(requestTo("http://test-api/completions"))
                .andExpect(method(HttpMethod.POST))
                .andExpect(header("Authorization", "Bearer test-api-key"))
                .andExpect(header("Content-Type", MediaType.APPLICATION_JSON_VALUE))
                .andRespond(withSuccess(expectedResponse, MediaType.APPLICATION_JSON));

        String result = deepseekClient.generateCompletion("test prompt");
        
        assertEquals(expectedResponse, result);
        mockServer.verify();
    }
} 