package com.careerup.authcore.model;

import jakarta.persistence.*;
import java.util.ArrayList;
import java.util.List;

/**
 * Represents one of the 5 domains assessed in the ILO test:
 * - LANG: Ngôn ngữ (Language)
 * - LOGIC: Phân tích - lôgic (Logic/Analysis)
 * - DESIGN: Hình học - màu sắc - thiết kế (Visual/Design)
 * - PEOPLE: Làm việc với con người (People)
 * - MECH: Thể chất - cơ khí (Physical/Mechanical)
 */
@Entity
@Table(name = "ilo_domains")
public class IloDomain {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true, length = 10)
    private String code;

    @Column(nullable = false)
    private String name;

    @Column(columnDefinition = "TEXT")
    private String description;

    @OneToMany(mappedBy = "domain")
    private List<IloQuestionDomain> questionDomains = new ArrayList<>();
    
    @OneToMany(mappedBy = "domain")
    private List<IloCareerMap> careerMappings = new ArrayList<>();

    // Getters and setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getCode() {
        return code;
    }

    public void setCode(String code) {
        this.code = code;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public List<IloQuestionDomain> getQuestionDomains() {
        return questionDomains;
    }

    public void setQuestionDomains(List<IloQuestionDomain> questionDomains) {
        this.questionDomains = questionDomains;
    }
    
    public List<IloCareerMap> getCareerMappings() {
        return careerMappings;
    }

    public void setCareerMappings(List<IloCareerMap> careerMappings) {
        this.careerMappings = careerMappings;
    }
}
