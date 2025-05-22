package com.careerup.authcore.model;

import jakarta.persistence.*;

/**
 * Maps domains to career fields for suggestions
 */
@Entity
@Table(name = "ilo_career_map")
public class IloCareerMap {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne
    @JoinColumn(name = "domain_id", nullable = false)
    private IloDomain domain;

    @Column(nullable = false)
    private String careerField;
    
    @Column(columnDefinition = "TEXT")
    private String description;
    
    @Column
    private Integer priority = 0;

    // Getters and setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public IloDomain getDomain() {
        return domain;
    }

    public void setDomain(IloDomain domain) {
        this.domain = domain;
    }

    public String getCareerField() {
        return careerField;
    }

    public void setCareerField(String careerField) {
        this.careerField = careerField;
    }
    
    public String getDescription() {
        return description;
    }
    
    public void setDescription(String description) {
        this.description = description;
    }
    
    public Integer getPriority() {
        return priority;
    }
    
    public void setPriority(Integer priority) {
        this.priority = priority;
    }
}
