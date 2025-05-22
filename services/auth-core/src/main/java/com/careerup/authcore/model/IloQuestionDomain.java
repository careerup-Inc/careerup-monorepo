package com.careerup.authcore.model;

import jakarta.persistence.*;

/**
 * Maps questions to domains
 */
@Entity
@Table(name = "ilo_question_domains")
public class IloQuestionDomain {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne
    @JoinColumn(name = "question_id", nullable = false)
    private IloQuestion question;

    @ManyToOne
    @JoinColumn(name = "domain_id", nullable = false)
    private IloDomain domain;

    @Column(name = "weight", nullable = false)
    private Double weight = 1.0; // Default weight

    // Getters and setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public IloQuestion getQuestion() {
        return question;
    }

    public void setQuestion(IloQuestion question) {
        this.question = question;
    }

    public IloDomain getDomain() {
        return domain;
    }

    public void setDomain(IloDomain domain) {
        this.domain = domain;
    }

    public Double getWeight() {
        return weight;
    }

    public void setWeight(Double weight) {
        this.weight = weight;
    }
}
