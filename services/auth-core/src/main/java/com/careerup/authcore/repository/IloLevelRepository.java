package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloLevel;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface IloLevelRepository extends JpaRepository<IloLevel, Long> {
    @Query("SELECT l FROM IloLevel l WHERE :percent BETWEEN l.minPercent AND l.maxPercent")
    Optional<IloLevel> findByPercentScore(Integer percent);
}
