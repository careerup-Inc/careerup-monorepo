package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloQuestionDomain;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface IloQuestionDomainRepository extends JpaRepository<IloQuestionDomain, Long> {
    List<IloQuestionDomain> findByQuestionId(Long questionId);
    
    @Query("SELECT qd FROM IloQuestionDomain qd WHERE qd.domain.code = :domainCode")
    List<IloQuestionDomain> findByDomainCode(String domainCode);
}
