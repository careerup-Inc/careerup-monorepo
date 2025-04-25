package com.careerup.llmgateway.model;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
@Schema(description = "Response containing the generated completion")
public class CompletionResponse {
    @Schema(description = "The generated text completion", example = "The capital of France is Paris.")
    private String completion;
    
    @Schema(description = "The model used for generation", example = "deepseek-r1")
    private String model;
    
    @Schema(description = "Number of tokens used in the completion", example = "42")
    private Long tokensUsed;
}