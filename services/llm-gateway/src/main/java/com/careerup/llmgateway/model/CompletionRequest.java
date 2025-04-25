package com.careerup.llmgateway.model;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.Data;

@Data
@Schema(description = "Request for text completion")
public class CompletionRequest {
    @Schema(description = "The prompt text to complete", required = true, example = "What is the capital of France?")
    private String prompt;
    
    @Schema(description = "Maximum number of tokens to generate", example = "1000")
    private Integer maxTokens;
    
    @Schema(description = "Sampling temperature (0.0 to 1.0)", example = "0.7")
    private Double temperature;
}