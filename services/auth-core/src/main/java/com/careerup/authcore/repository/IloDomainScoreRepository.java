package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloDomainScore;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface IloDomainScoreRepository extends JpaRepository<IloDomainScore, Long> {
    List<IloDomainScore> findByTestResultIdOrderByRankAsc(Long testResultId);
}
