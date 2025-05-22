package com.careerup.authcore.service;

import com.careerup.authcore.model.IloDomainScore;
import com.careerup.authcore.model.IloTestResult;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * Service that integrates with LLM gateway for personalized insights on ILO test results
 */
@Service
public class LlmIntegrationService {

    @Autowired
    private RestTemplate restTemplate;

    @Value("${llm.gateway.url:http://llm-gateway:8080}")
    private String llmGatewayUrl;

    /**
     * Generate personalized insights for ILO test results
     * @param testResult The ILO test result to analyze
     * @return Personalized insights based on LLM analysis
     */
    public String generatePersonalizedInsights(IloTestResult testResult) {
        // Create a rich prompt for LLM analysis
        String prompt = buildPromptFromTestResult(testResult);
        
        // Call LLM Gateway API
        Map<String, Object> requestBody = new HashMap<>();
        requestBody.put("prompt", prompt);
        requestBody.put("user_id", testResult.getUserId().toString());
        requestBody.put("max_tokens", 1500);
        
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        HttpEntity<Map<String, Object>> request = new HttpEntity<>(requestBody, headers);
        
        try {
            @SuppressWarnings("unchecked")
            Map<String, Object> response = (Map<String, Object>) restTemplate.postForObject(
                llmGatewayUrl + "/api/v1/generate",
                request,
                Map.class
            );
            
            if (response != null && response.containsKey("text")) {
                return (String) response.get("text");
            } else {
                return "Không thể tạo phân tích cá nhân hoá. Vui lòng thử lại sau.";
            }
        } catch (Exception e) {
            return "Đã xảy ra lỗi khi phân tích kết quả: " + e.getMessage();
        }
    }
    
    /**
     * Build a detailed prompt for LLM analysis based on the test result
     * @param testResult The test result to analyze
     * @return A prompt for the LLM
     */
    private String buildPromptFromTestResult(IloTestResult testResult) {
        StringBuilder promptBuilder = new StringBuilder();
        promptBuilder.append("Phân tích kết quả bài kiểm tra hướng nghiệp ILO với các điểm số theo từng lĩnh vực sau:\n\n");
        
        // Add domain scores
        List<IloDomainScore> sortedScores = testResult.getDomainScores().stream()
                .sorted((a, b) -> Float.compare(b.getPercentScore(), a.getPercentScore()))
                .collect(Collectors.toList());
        
        for (IloDomainScore score : sortedScores) {
            promptBuilder.append("- ")
                    .append(score.getDomain().getName())
                    .append(" (").append(score.getDomain().getCode()).append("): ")
                    .append(String.format("%.1f", score.getPercentScore())).append("% ")
                    .append("(").append(score.getLevel().getLevelName()).append(")\n");
        }
        
        // Add suggested careers
        promptBuilder.append("\nNghề nghiệp được đề xuất: ");
        if (testResult.getSuggestedCareers() != null && !testResult.getSuggestedCareers().isEmpty()) {
            promptBuilder.append(testResult.getSuggestedCareers());
        } else {
            promptBuilder.append("Không có đề xuất");
        }
        
        // Instructions for LLM
        promptBuilder.append("\n\nHãy phân tích kết quả trên và đưa ra nhận xét cá nhân hóa bằng tiếng Việt. Phân tích nên bao gồm:");
        promptBuilder.append("\n1. Điểm mạnh và điểm yếu của người làm bài theo các lĩnh vực");
        promptBuilder.append("\n2. Gợi ý về hướng phát triển nghề nghiệp phù hợp dựa trên điểm mạnh");
        promptBuilder.append("\n3. Gợi ý về cách phát triển kỹ năng ở các lĩnh vực yếu hơn");
        promptBuilder.append("\n4. Đề xuất 3-5 nghề nghiệp cụ thể phù hợp với điểm mạnh, giải thích ngắn gọn lý do");
        
        return promptBuilder.toString();
    }
}
