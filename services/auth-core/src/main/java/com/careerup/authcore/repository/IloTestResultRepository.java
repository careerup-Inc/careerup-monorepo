package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloTestResult;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface IloTestResultRepository extends JpaRepository<IloTestResult, Long> {
    List<IloTestResult> findByUserId(UUID userId);
    
    // Order by createdAt descending to get the most recent results first
    List<IloTestResult> findByUserIdOrderByCreatedAtDesc(UUID userId);
    
    @Query("SELECT r FROM IloTestResult r LEFT JOIN FETCH r.domainScores WHERE r.id = :id")
    IloTestResult findByIdWithDomainScores(Long id);
}
