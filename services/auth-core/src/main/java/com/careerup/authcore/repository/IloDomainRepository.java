package com.careerup.authcore.repository;

import com.careerup.authcore.model.IloDomain;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface IloDomainRepository extends JpaRepository<IloDomain, Long> {
    Optional<IloDomain> findByCode(String code);
}
