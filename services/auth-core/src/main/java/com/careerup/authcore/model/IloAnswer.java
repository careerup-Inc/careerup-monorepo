package com.careerup.authcore.model;

import jakarta.persistence.*;
import java.time.LocalDateTime;
import java.util.UUID;

/**
 * Represents a single answer to an ILO test question
 */
@Entity
@Table(name = "ilo_answers")
public class IloAnswer {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private UUID userId;

    @ManyToOne
    @JoinColumn(name = "question_id", nullable = false)
    private IloQuestion question;

    @Column(nullable = false)
    private Integer selectedOption;

    @ManyToOne
    @JoinColumn(name = "test_result_id", nullable = false)
    private IloTestResult testResult;

    @Column(nullable = false)
    private LocalDateTime createdAt = LocalDateTime.now();

    // Getters and setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public UUID getUserId() {
        return userId;
    }

    public void setUserId(UUID userId) {
        this.userId = userId;
    }

    public IloQuestion getQuestion() {
        return question;
    }

    public void setQuestion(IloQuestion question) {
        this.question = question;
    }

    public Integer getSelectedOption() {
        return selectedOption;
    }

    public void setSelectedOption(Integer selectedOption) {
        this.selectedOption = selectedOption;
    }

    public IloTestResult getTestResult() {
        return testResult;
    }

    public void setTestResult(IloTestResult testResult) {
        this.testResult = testResult;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }
}
