package com.careerup.authcore.service;

import com.careerup.authcore.model.*;
import com.careerup.authcore.repository.*;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class IloTestResultService {
    private final IloTestResultRepository iloTestResultRepository;
    private final IloQuestionRepository iloQuestionRepository;
    private final IloAnswerRepository iloAnswerRepository;
    private final IloDomainService iloDomainService;

    /**
     * Legacy method for backward compatibility
     */
    @Transactional
    public IloTestResult saveResult(UUID userId, String resultData) {
        IloTestResult result = new IloTestResult();
        result.setUserId(userId);
        result.setResultData(resultData);
        return iloTestResultRepository.save(result);
    }

    /**
     * Get all test results for a specific user
     */
    @Transactional(readOnly = true)
    public List<IloTestResult> getResultsByUserId(UUID userId) {
        List<IloTestResult> results = iloTestResultRepository.findByUserIdOrderByCreatedAtDesc(userId);
        // Initialize lazy collections to avoid LazyInitializationException
        results.forEach(result -> {
            if (result.getDomainScores() != null) {
                result.getDomainScores().size(); // Force initialization
                result.getDomainScores().forEach(score -> {
                    if (score.getDomain() != null) {
                        score.getDomain().getCode(); // Force initialization
                    }
                    if (score.getLevel() != null) {
                        score.getLevel().getLevelName(); // Force initialization
                    }
                });
            }
        });
        return results;
    }
    
    /**
     * Get a specific test result by ID
     */
    @Transactional(readOnly = true)
    public Optional<IloTestResult> getResultById(Long id) {
        IloTestResult result = iloTestResultRepository.findByIdWithDomainScores(id);
        if (result != null && result.getDomainScores() != null) {
            // Initialize lazy collections to avoid LazyInitializationException
            result.getDomainScores().size(); // Force initialization
            result.getDomainScores().forEach(score -> {
                if (score.getDomain() != null) {
                    score.getDomain().getCode(); // Force initialization
                }
                if (score.getLevel() != null) {
                    score.getLevel().getLevelName(); // Force initialization
                }
            });
        }
        return Optional.ofNullable(result);
    }
    
    /**
     * Save a test result with structured answers and calculate scores
     */
    @Transactional
    public IloTestResult saveResultWithAnswers(UUID userId, String resultData, List<com.careerup.proto.v1.IloAnswer> protoAnswers) {
        // Invalidate cache for this user before saving new result
        iloDomainService.invalidateUserResultCache(userId.toString());
        
        // Create the test result
        IloTestResult result = new IloTestResult();
        result.setUserId(userId);
        result.setResultData(resultData); // Keep raw data for backward compatibility
        result.setCreatedAt(java.time.LocalDateTime.now());
        
        // Save the result first to get an ID
        result = iloTestResultRepository.save(result);
        
        // Process answers if provided
        if (protoAnswers != null && !protoAnswers.isEmpty()) {
            List<com.careerup.authcore.model.IloAnswer> answers = convertAndSaveAnswers(userId, result, protoAnswers);
            
            // Calculate domain scores
            List<IloDomainScore> domainScores = iloDomainService.calculateDomainScores(answers);
            
            // Attach scores to result
            for (IloDomainScore score : domainScores) {
                result.addDomainScore(score);
            }
            
            // Get career suggestions based on top domains (limit to 5)
            List<IloCareerMap> careerSuggestions = iloDomainService.getCareerSuggestions(domainScores, 5);
            
            // Store suggested careers as comma-separated list
            if (!careerSuggestions.isEmpty()) {
                String careers = careerSuggestions.stream()
                    .map(IloCareerMap::getCareerField)
                    .collect(Collectors.joining(","));
                result.setSuggestedCareers(careers);
            }
            
            // Update the result with scores and suggestions
            result = iloTestResultRepository.save(result);
        }
        
        return result;
    }
    
    /**
     * Convert proto answers to entity answers and save them
     */
    private List<com.careerup.authcore.model.IloAnswer> convertAndSaveAnswers(UUID userId, IloTestResult result, List<com.careerup.proto.v1.IloAnswer> protoAnswers) {
        List<com.careerup.authcore.model.IloAnswer> answers = new ArrayList<>();
        
        for (com.careerup.proto.v1.IloAnswer pa : protoAnswers) {
            // Find the question by ID
            Optional<IloQuestion> questionOpt = iloQuestionRepository.findById(Long.parseLong(pa.getQuestionId()));
            if (questionOpt.isEmpty()) {
                continue; // Skip if question not found
            }
            
            IloQuestion question = questionOpt.get();
            
            // Create and save the answer
            com.careerup.authcore.model.IloAnswer answer = new com.careerup.authcore.model.IloAnswer();
            answer.setUserId(userId);
            answer.setQuestion(question);
            answer.setSelectedOption(pa.getSelectedOption());
            answer.setTestResult(result);
            
            answers.add(answer);
        }
        
        // Save all answers at once
        return iloAnswerRepository.saveAll(answers);
    }
}
