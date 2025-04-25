package com.careerup.llmgateway.controller;

import com.careerup.llmgateway.model.CompletionRequest;
import com.careerup.llmgateway.model.CompletionResponse;
import com.careerup.llmgateway.service.DeepseekService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@Tag(name = "Deepseek API", description = "Operations related to the Deepseek AI model")
@RestController
@RequestMapping("/api/v1/deepseek")
@RequiredArgsConstructor
public class DeepseekController {
    private final DeepseekService deepseekService;

    @Operation(summary = "Generate text completion using Deepseek model")
    @ApiResponses(value = {
        @ApiResponse(responseCode = "200", description = "Successfully generated completion",
                    content = @Content(mediaType = "application/json", 
                    schema = @Schema(implementation = CompletionResponse.class))),
        @ApiResponse(responseCode = "500", description = "Internal server error",
                    content = @Content)
    })
    @PostMapping("/completion")
    public CompletionResponse generateCompletion(@RequestBody CompletionRequest request) {
        return deepseekService.processPrompt(request);
    }
}