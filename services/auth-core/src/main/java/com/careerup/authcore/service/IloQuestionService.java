package com.careerup.authcore.service;

import com.careerup.authcore.model.IloQuestion;
import com.careerup.authcore.repository.IloQuestionRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.*;

@Service
public class IloQuestionService {
    private final IloQuestionRepository iloQuestionRepository;

    public IloQuestionService(IloQuestionRepository iloQuestionRepository) {
        this.iloQuestionRepository = iloQuestionRepository;
    }

    /**
     * Get all ILO questions with their scales
     */
    @Transactional(readOnly = true)
    public List<IloQuestion> getAllQuestions() {
        List<IloQuestion> questions = iloQuestionRepository.findAllByOrderByQuestionNumberAsc();
        System.out.println("IloQuestionService: Retrieved " + questions.size() + " questions from repository");
        
        // Check if questions have domains
        long questionsWithoutDomains = questions.stream()
            .filter(q -> q.getQuestionDomains() == null || q.getQuestionDomains().isEmpty())
            .count();
        
        if (questionsWithoutDomains > 0) {
            System.out.println("Warning: " + questionsWithoutDomains + " questions don't have domain mappings!");
        }
        if (!questions.isEmpty()) {
            // Print domain distribution
            Map<String, Integer> domainCounts = new HashMap<>();
            for (IloQuestion question : questions) {
                if (!question.getQuestionDomains().isEmpty()) {
                    String domainCode = question.getQuestionDomains().get(0).getDomain().getCode();
                    domainCounts.put(domainCode, domainCounts.getOrDefault(domainCode, 0) + 1);
                }
            }
            System.out.println("Domain distribution: " + domainCounts);
        }
        return questions;
    }

    /**
     * Find a question by its ID
     */
    public Optional<IloQuestion> getQuestionById(Long id) {
        return iloQuestionRepository.findById(id);
    }

    /**
     * Save a new ILO question
     */
    @Transactional
    public IloQuestion saveQuestion(IloQuestion question) {
        return iloQuestionRepository.save(question);
    }

    /**
     * Save a batch of questions
     */
    @Transactional
    public List<IloQuestion> saveAllQuestions(List<IloQuestion> questions) {
        return iloQuestionRepository.saveAll(questions);
    }
}
