package com.careerup.authcore.model;

import jakarta.persistence.*;
import java.util.ArrayList;
import java.util.List;

@Entity
@Table(name = "ilo_questions")
public class IloQuestion {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String questionText;

    @Column(nullable = false)
    private Integer questionNumber;

    // Options are stored as a comma-separated list to avoid creating another table
    // For ILO, options are typically the same across all questions (e.g., 1-4 rating)
    @ElementCollection(fetch = FetchType.EAGER)
    @CollectionTable(name = "ilo_question_options", joinColumns = @JoinColumn(name = "question_id"))
    @Column(name = "option_text")
    private List<String> options = new ArrayList<>();
    
    // One question can map to multiple domains with different weights
    @OneToMany(mappedBy = "question", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<IloQuestionDomain> questionDomains = new ArrayList<>();
    
    // One question can have multiple answers from different users
    @OneToMany(mappedBy = "question")
    private List<IloAnswer> answers = new ArrayList<>();

    // Getters and setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getQuestionText() {
        return questionText;
    }

    public void setQuestionText(String questionText) {
        this.questionText = questionText;
    }

    public Integer getQuestionNumber() {
        return questionNumber;
    }

    public void setQuestionNumber(Integer questionNumber) {
        this.questionNumber = questionNumber;
    }

    public List<String> getOptions() {
        return options;
    }

    public void setOptions(List<String> options) {
        this.options = options;
    }
    
    public List<IloQuestionDomain> getQuestionDomains() {
        return questionDomains;
    }
    
    public void setQuestionDomains(List<IloQuestionDomain> questionDomains) {
        this.questionDomains = questionDomains;
    }
    
    public void addQuestionDomain(IloDomain domain, Double weight) {
        IloQuestionDomain questionDomain = new IloQuestionDomain();
        questionDomain.setQuestion(this);
        questionDomain.setDomain(domain);
        questionDomain.setWeight(weight);
        this.questionDomains.add(questionDomain);
    }
    
    public List<IloAnswer> getAnswers() {
        return answers;
    }
    
    public void setAnswers(List<IloAnswer> answers) {
        this.answers = answers;
    }
}
