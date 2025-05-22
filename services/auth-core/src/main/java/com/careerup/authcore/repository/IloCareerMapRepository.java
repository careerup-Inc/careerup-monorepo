package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloCareerMap;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface IloCareerMapRepository extends JpaRepository<IloCareerMap, Long> {
    List<IloCareerMap> findByDomainCode(String domainCode);
    
    @Query("SELECT c FROM IloCareerMap c WHERE c.domain.code IN :domainCodes ORDER BY c.priority DESC")
    List<IloCareerMap> findByDomainCodesOrderedByPriority(List<String> domainCodes);
}
