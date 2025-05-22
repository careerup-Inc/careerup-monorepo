package com.careerup.authcore.model;

import jakarta.persistence.*;

/**
 * Represents a scored domain for a user's test result
 */
@Entity
@Table(name = "ilo_domain_scores")
public class IloDomainScore {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne
    @JoinColumn(name = "test_result_id", nullable = false)
    private IloTestResult testResult;

    @ManyToOne
    @JoinColumn(name = "domain_id", nullable = false)
    private IloDomain domain;

    @Column(nullable = false)
    private Integer rawScore;

    @Column(nullable = false)
    private Float percentScore;

    @ManyToOne
    @JoinColumn(name = "level_id")
    private IloLevel level;

    @Column
    private Integer rank;

    // Getters and setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public IloTestResult getTestResult() {
        return testResult;
    }

    public void setTestResult(IloTestResult testResult) {
        this.testResult = testResult;
    }

    public IloDomain getDomain() {
        return domain;
    }

    public void setDomain(IloDomain domain) {
        this.domain = domain;
    }

    public Integer getRawScore() {
        return rawScore;
    }

    public void setRawScore(Integer rawScore) {
        this.rawScore = rawScore;
    }

    public Float getPercentScore() {
        return percentScore;
    }

    public void setPercentScore(Float percentScore) {
        this.percentScore = percentScore;
    }

    public IloLevel getLevel() {
        return level;
    }

    public void setLevel(IloLevel level) {
        this.level = level;
    }

    public Integer getRank() {
        return rank;
    }

    public void setRank(Integer rank) {
        this.rank = rank;
    }
}
