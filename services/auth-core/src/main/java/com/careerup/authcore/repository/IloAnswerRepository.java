package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloAnswer;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.UUID;

@Repository
public interface IloAnswerRepository extends JpaRepository<IloAnswer, Long> {
    List<IloAnswer> findByTestResultId(Long testResultId);
    List<IloAnswer> findByUserId(UUID userId);
}
