package com.careerup.llmgateway.model;

import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.Data;

@Data
@ApiModel(description = "Request for text completion")
public class CompletionRequest {
    @ApiModelProperty(value = "The prompt text to complete", required = true, example = "What is the capital of France?")
    private String prompt;
    
    @ApiModelProperty(value = "Maximum number of tokens to generate", example = "1000")
    private Integer maxTokens;
    
    @ApiModelProperty(value = "Sampling temperature (0.0 to 1.0)", example = "0.7")
    private Double temperature;
} 