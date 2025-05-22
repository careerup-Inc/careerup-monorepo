package com.careerup.authcore.controller;

import com.careerup.authcore.model.IloCareerMap;
import com.careerup.authcore.model.IloDomain;
import com.careerup.authcore.model.IloLevel;
import com.careerup.authcore.model.IloQuestion;
import com.careerup.authcore.service.IloDomainService;
import com.careerup.authcore.service.IloQuestionService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.Optional;

/**
 * REST controller for managing ILO test components through an admin interface
 */
@RestController
@RequestMapping("/api/v1/admin/ilo")
public class IloAdminController {

    // TODO review this
    @Autowired
    private IloDomainService iloDomainService;

    @Autowired
    private IloQuestionService iloQuestionService;

    // Domain management endpoints
    
    @GetMapping("/domains")
    public ResponseEntity<List<IloDomain>> getAllDomains() {
        return ResponseEntity.ok(iloDomainService.getAllDomains());
    }
    
    @PostMapping("/domains")
    public ResponseEntity<IloDomain> createDomain(@RequestBody IloDomain domain) {
        // Implementation would add validation and save the domain
        // This is a placeholder until actual repository method is implemented
        return ResponseEntity.status(HttpStatus.CREATED).body(domain);
    }
    
    @PutMapping("/domains/{id}")
    public ResponseEntity<IloDomain> updateDomain(@PathVariable Long id, @RequestBody IloDomain domain) {
        // Implementation would validate, find by ID, update and save
        domain.setId(id);
        return ResponseEntity.ok(domain);
    }
    
    // Level management endpoints
    
    @GetMapping("/levels")
    public ResponseEntity<List<IloLevel>> getAllLevels() {
        return ResponseEntity.ok(iloDomainService.getAllLevels());
    }
    
    @PostMapping("/levels")
    public ResponseEntity<IloLevel> createLevel(@RequestBody IloLevel level) {
        // Implementation would validate the level ranges don't overlap and save it
        return ResponseEntity.status(HttpStatus.CREATED).body(level);
    }
    
    @PutMapping("/levels/{id}")
    public ResponseEntity<IloLevel> updateLevel(@PathVariable Long id, @RequestBody IloLevel level) {
        level.setId(id);
        return ResponseEntity.ok(level);
    }
    
    // Career mapping endpoints
    
    @GetMapping("/careers")
    public ResponseEntity<Map<String, List<IloCareerMap>>> getAllCareerMappingsByDomain() {
        // Get all domains
        List<IloDomain> domains = iloDomainService.getAllDomains();
        
        // Create dummy domain scores (100% for each) to get all career maps
        List<IloCareerMap> allCareers = iloDomainService.getCareerSuggestions(null, 100);
        
        // Group by domain (you'll need to implement this method)
        Map<String, List<IloCareerMap>> careersByDomain = iloDomainService.getGroupedCareerSuggestions(null, 100, domains.size());
        
        return ResponseEntity.ok(careersByDomain);
    }
    
    @PostMapping("/careers")
    public ResponseEntity<IloCareerMap> createCareerMapping(@RequestBody IloCareerMap careerMap) {
        // Implementation would validate and save
        return ResponseEntity.status(HttpStatus.CREATED).body(careerMap);
    }
    
    @PutMapping("/careers/{id}")
    public ResponseEntity<IloCareerMap> updateCareerMapping(@PathVariable Long id, @RequestBody IloCareerMap careerMap) {
        careerMap.setId(id);
        return ResponseEntity.ok(careerMap);
    }
    
    // Question management endpoints
    
    @GetMapping("/questions")
    public ResponseEntity<List<IloQuestion>> getAllQuestions() {
        return ResponseEntity.ok(iloQuestionService.getAllQuestions());
    }
    
    @GetMapping("/questions/{id}")
    public ResponseEntity<IloQuestion> getQuestion(@PathVariable Long id) {
        Optional<IloQuestion> question = iloQuestionService.getQuestionById(id);
        return question.map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
    
    @PostMapping("/questions")
    public ResponseEntity<IloQuestion> createQuestion(@RequestBody IloQuestion question) {
        IloQuestion saved = iloQuestionService.saveQuestion(question);
        return ResponseEntity.status(HttpStatus.CREATED).body(saved);
    }
    
    @PutMapping("/questions/{id}")
    public ResponseEntity<IloQuestion> updateQuestion(@PathVariable Long id, @RequestBody IloQuestion question) {
        question.setId(id);
        IloQuestion updated = iloQuestionService.saveQuestion(question);
        return ResponseEntity.ok(updated);
    }
    
    // Bulk operations
    
    @PostMapping("/initialize")
    public ResponseEntity<String> initializeAllData() {
        // Initialize all default data
        iloDomainService.initializeDefaultDomains();
        iloDomainService.initializeDefaultLevels();
        iloDomainService.initializeDefaultCareerMappings();
        return ResponseEntity.ok("All ILO test data has been initialized successfully");
    }
}
