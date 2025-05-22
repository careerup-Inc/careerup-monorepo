package com.careerup.authcore.model;

import jakarta.persistence.*;

/**
 * Represents the evaluation level for a domain score:
 * - Rất mạnh (Very Strong): ≥80%
 * - Mạnh (Strong): 60-79%
 * - Trung bình (Average): 40-59%
 * - Yếu (Weak): <40%
 */
@Entity
@Table(name = "ilo_levels")
public class IloLevel {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private Integer minPercent;

    @Column(nullable = false)
    private Integer maxPercent;

    @Column(nullable = false)
    private String levelName;

    @Column(columnDefinition = "TEXT")
    private String suggestion;

    // Getters and setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Integer getMinPercent() {
        return minPercent;
    }

    public void setMinPercent(Integer minPercent) {
        this.minPercent = minPercent;
    }

    public Integer getMaxPercent() {
        return maxPercent;
    }

    public void setMaxPercent(Integer maxPercent) {
        this.maxPercent = maxPercent;
    }

    public String getLevelName() {
        return levelName;
    }

    public void setLevelName(String levelName) {
        this.levelName = levelName;
    }

    public String getSuggestion() {
        return suggestion;
    }

    public void setSuggestion(String suggestion) {
        this.suggestion = suggestion;
    }
}
