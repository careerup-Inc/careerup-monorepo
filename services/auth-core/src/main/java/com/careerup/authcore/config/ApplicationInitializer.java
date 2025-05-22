package com.careerup.authcore.config;

import com.careerup.authcore.service.IloDomainService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

/**
 * Initializes application data when the application starts
 */
@Component
public class ApplicationInitializer implements CommandLineRunner {

    @Autowired
    private IloDomainService iloDomainService;
    
    @Override
    public void run(String... args) {
        // Initialize default domains and levels
        iloDomainService.initializeDefaultDomains();
        iloDomainService.initializeDefaultLevels();
        // Initialize default career mappings
        iloDomainService.initializeDefaultCareerMappings();
    }
}
