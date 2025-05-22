package com.careerup.authcore.model;

import jakarta.persistence.*;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@Entity
@Table(name = "ilo_test_results")
public class IloTestResult {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private UUID userId;

    @Column(nullable = false, columnDefinition = "TEXT")
    private String resultData;

    @Column(nullable = false)
    private LocalDateTime createdAt = LocalDateTime.now();
    
    @OneToMany(mappedBy = "testResult", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<IloDomainScore> domainScores = new ArrayList<>();
    
    @OneToMany(mappedBy = "testResult", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<IloAnswer> answers = new ArrayList<>();
    
    @Column(columnDefinition = "TEXT")
    private String suggestedCareers;

    // Getters and setters
    public Long getId() { return id; }
    public void setId(Long id) { this.id = id; }

    public UUID getUserId() { return userId; }
    public void setUserId(UUID userId) { this.userId = userId; }

    public String getResultData() { return resultData; }
    public void setResultData(String resultData) { this.resultData = resultData; }

    public LocalDateTime getCreatedAt() { return createdAt; }
    public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }
    
    public List<IloDomainScore> getDomainScores() {
        return domainScores;
    }
    
    public void setDomainScores(List<IloDomainScore> domainScores) {
        this.domainScores = domainScores;
    }
    
    public void addDomainScore(IloDomainScore domainScore) {
        domainScores.add(domainScore);
        domainScore.setTestResult(this);
    }
    
    public List<IloAnswer> getAnswers() {
        return answers;
    }
    
    public void setAnswers(List<IloAnswer> answers) {
        this.answers = answers;
    }
    
    public void addAnswer(IloAnswer answer) {
        answers.add(answer);
        answer.setTestResult(this);
    }
    
    public String getSuggestedCareers() {
        return suggestedCareers;
    }
    
    public void setSuggestedCareers(String suggestedCareers) {
        this.suggestedCareers = suggestedCareers;
    }
}
