package com.careerup.llmgateway.controller;

import com.careerup.llmgateway.model.CompletionRequest;
import com.careerup.llmgateway.model.CompletionResponse;
import com.careerup.llmgateway.service.DeepseekService;
import io.swagger.annotations.Api;
import io.swagger.annotations.ApiOperation;
import io.swagger.annotations.ApiResponse;
import io.swagger.annotations.ApiResponses;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@Api(tags = "Deepseek API")
@RestController
@RequestMapping("/api/v1/deepseek")
@RequiredArgsConstructor
public class DeepseekController {
    private final DeepseekService deepseekService;

    @ApiOperation(value = "Generate text completion using Deepseek model")
    @ApiResponses(value = {
        @ApiResponse(code = 200, message = "Successfully generated completion"),
        @ApiResponse(code = 500, message = "Internal server error")
    })
    @PostMapping("/completion")
    public CompletionResponse generateCompletion(@RequestBody CompletionRequest request) {
        return deepseekService.processPrompt(request);
    }
} 