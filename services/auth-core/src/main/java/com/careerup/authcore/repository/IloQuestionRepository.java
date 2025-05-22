package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloQuestion;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface IloQuestionRepository extends JpaRepository<IloQuestion, Long> {
    List<IloQuestion> findAllByOrderByQuestionNumberAsc();
}
