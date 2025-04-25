package com.careerup.llmgateway.model;

import io.swagger.annotations.ApiModel;
import io.swagger.annotations.ApiModelProperty;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@ApiModel(description = "Response containing the generated completion")
public class CompletionResponse {
    @ApiModelProperty(value = "The generated text completion", example = "The capital of France is Paris.")
    private String completion;
    
    @ApiModelProperty(value = "The model used for generation", example = "deepseek-r1")
    private String model;
    
    @ApiModelProperty(value = "Number of tokens used in the completion", example = "42")
    private Long tokensUsed;
} 