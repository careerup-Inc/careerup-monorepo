package com.careerup.authcore.service;

import com.careerup.authcore.model.*;
import com.careerup.authcore.repository.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Service
public class IloDomainService {
    
    @Autowired
    private IloDomainRepository domainRepository;
    
    @Autowired
    private IloLevelRepository levelRepository;
    
    @Autowired
    private IloQuestionDomainRepository questionDomainRepository;
    
    @Autowired
    private IloCareerMapRepository careerMapRepository;

    @Autowired
    private RedisTemplate<String, Object> redisTemplate;

    @Value("${ilo.cache.expiration.seconds:1800}")
    private long cacheExpirationSeconds;

    private String getUserResultCacheKey(String userId) {
        return "ilo:user:result:" + userId;
    }

    public void cacheUserResult(String userId, Object result) {
        String key = getUserResultCacheKey(userId);
        redisTemplate.opsForValue().set(key, result, cacheExpirationSeconds, TimeUnit.SECONDS);
    }

    public Object getCachedUserResult(String userId) {
        String key = getUserResultCacheKey(userId);
        return redisTemplate.opsForValue().get(key);
    }

    public void invalidateUserResultCache(String userId) {
        String key = getUserResultCacheKey(userId);
        redisTemplate.delete(key);
    }

    /**
     * Get all available domains
     */
    public List<IloDomain> getAllDomains() {
        return domainRepository.findAll();
    }

    /**
     * Get all available evaluation levels
     */
    public List<IloLevel> getAllLevels() {
        return levelRepository.findAll();
    }
    
    /**
     * Find the appropriate level for a given percentage score
     */
    public IloLevel findLevelForScore(Integer percentScore) {
        return levelRepository.findByPercentScore(percentScore)
                .orElseThrow(() -> new RuntimeException("Invalid percent score: " + percentScore));
    }
    
    /**
     * Calculate domain scores for a set of answers
     * @param answers The user's answers to ILO test questions
     * @return List of domain scores
     */
    @Transactional
    public List<IloDomainScore> calculateDomainScores(List<IloAnswer> answers) {
        // Group questions by domain
        Map<String, List<IloQuestionDomain>> domainMap = getAllQuestionDomainMapping();
        
        // Calculate raw scores for each domain
        Map<String, Double> rawScores = calculateRawScores(answers, domainMap);
        
        // Convert to domain scores with levels
        List<IloDomainScore> scores = new ArrayList<>();
        for (Map.Entry<String, Double> entry : rawScores.entrySet()) {
            String domainCode = entry.getKey();
            double rawScore = entry.getValue();
            
            // Maximum score for a domain is 48
            float percentScore = Math.min(100, (float) (rawScore / 48.0 * 100));
            
            IloDomain domain = domainRepository.findByCode(domainCode)
                    .orElseThrow(() -> new RuntimeException("Domain not found: " + domainCode));
            
            IloLevel level = findLevelForScore((int) percentScore);
            
            IloDomainScore score = new IloDomainScore();
            score.setDomain(domain);
            score.setRawScore((int) rawScore);
            score.setPercentScore(percentScore);
            score.setLevel(level);
            
            scores.add(score);
        }
        
        // Sort by percent score descending and assign ranks
        scores.sort(Comparator.comparing(IloDomainScore::getPercentScore).reversed());
        for (int i = 0; i < scores.size(); i++) {
            scores.get(i).setRank(i + 1);
        }
        
        return scores;
    }
    
    /**
     * Calculate raw scores for each domain based on user answers
     */
    private Map<String, Double> calculateRawScores(List<IloAnswer> answers, Map<String, List<IloQuestionDomain>> domainMap) {
        // Initialize scores for each domain to 0
        Map<String, Double> scores = domainMap.keySet().stream()
                .collect(Collectors.toMap(domain -> domain, domain -> 0.0));
        
        // Process each answer
        for (IloAnswer answer : answers) {
            IloQuestion question = answer.getQuestion();
            List<IloQuestionDomain> domains = question.getQuestionDomains();
            
            // For each domain this question belongs to
            for (IloQuestionDomain qd : domains) {
                String domainCode = qd.getDomain().getCode();
                double weight = qd.getWeight();
                
                // Add the weighted score to the domain
                double currentScore = scores.getOrDefault(domainCode, 0.0);
                scores.put(domainCode, currentScore + (answer.getSelectedOption() * weight));
            }
        }
        
        return scores;
    }
    
    /**
     * Get mapping of all questions to their domains
     */
    private Map<String, List<IloQuestionDomain>> getAllQuestionDomainMapping() {
        // Get all question-domain mappings
        List<IloQuestionDomain> allMappings = questionDomainRepository.findAll();
        
        // Group by domain code
        return allMappings.stream()
                .collect(Collectors.groupingBy(qd -> qd.getDomain().getCode()));
    }
    
    /**
     * Get career suggestions based on top domain scores
     */
    @Transactional
    public List<IloCareerMap> getCareerSuggestions(List<IloDomainScore> scores, int maxSuggestions) {
        // Get top 2 domains
        List<String> topDomainCodes = scores.stream()
                .sorted(Comparator.comparing(IloDomainScore::getPercentScore).reversed())
                .limit(2)
                .map(score -> score.getDomain().getCode())
                .collect(Collectors.toList());
        
        // Get career suggestions for these domains
        List<IloCareerMap> suggestions = careerMapRepository.findByDomainCodesOrderedByPriority(topDomainCodes);
        
        // Limit the number of suggestions if needed
        if (suggestions.size() > maxSuggestions) {
            suggestions = suggestions.subList(0, maxSuggestions);
        }
        
        return suggestions;
    }
    
    /**
     * Get career suggestions based on top domain scores with grouping by domain
     * @param scores List of domain scores
     * @param maxPerDomain Maximum suggestions per domain
     * @param topDomains Number of top domains to consider
     * @return Map of domain codes to career suggestions
     */
    @Transactional
    public Map<String, List<IloCareerMap>> getGroupedCareerSuggestions(
            List<IloDomainScore> scores, int maxPerDomain, int topDomains) {
        
        // Get top N domains
        List<IloDomainScore> topDomainScores = scores.stream()
                .sorted(Comparator.comparing(IloDomainScore::getPercentScore).reversed())
                .limit(topDomains)
                .collect(Collectors.toList());
        
        Map<String, List<IloCareerMap>> result = new HashMap<>();
        
        // For each top domain, get career suggestions
        for (IloDomainScore score : topDomainScores) {
            String domainCode = score.getDomain().getCode();
            List<IloCareerMap> careers = careerMapRepository.findByDomainCode(domainCode);
            
            // Sort by priority and limit the number
            List<IloCareerMap> topCareers = careers.stream()
                    .sorted(Comparator.comparing(IloCareerMap::getPriority).reversed())
                    .limit(maxPerDomain)
                    .collect(Collectors.toList());
            
            result.put(domainCode, topCareers);
        }
        
        return result;
    }
    
    /**
     * Initialize default domains if they don't exist
     */
    @Transactional
    public void initializeDefaultDomains() {
        if (domainRepository.count() > 0) {
            // Domains already exist
            return;
        }
        
        // Create the 5 Vietnamese ILO domains
        List<IloDomain> domains = new ArrayList<>();
        
        IloDomain lang = new IloDomain();
        lang.setCode("LANG");
        lang.setName("Ngôn ngữ");
        lang.setDescription("Khả năng học và sử dụng ngôn ngữ, kỹ năng giao tiếp");
        domains.add(lang);
        
        IloDomain logic = new IloDomain();
        logic.setCode("LOGIC");
        logic.setName("Phân tích - lôgic");
        logic.setDescription("Khả năng tư duy logic, giải quyết vấn đề, phân tích số liệu");
        domains.add(logic);
        
        IloDomain design = new IloDomain();
        design.setCode("DESIGN");
        design.setName("Hình học - màu sắc - thiết kế");
        design.setDescription("Khả năng thẩm mỹ, sáng tạo, nhận biết không gian và màu sắc");
        domains.add(design);
        
        IloDomain people = new IloDomain();
        people.setCode("PEOPLE");
        people.setName("Làm việc với con người");
        people.setDescription("Khả năng giao tiếp, đồng cảm, lãnh đạo và làm việc nhóm");
        domains.add(people);
        
        IloDomain mech = new IloDomain();
        mech.setCode("MECH");
        mech.setName("Thể chất - cơ khí");
        mech.setDescription("Khả năng vận động, thao tác, làm việc với máy móc, công cụ");
        domains.add(mech);
        
        domainRepository.saveAll(domains);
    }
    
    /**
     * Initialize default levels if they don't exist
     */
    @Transactional
    public void initializeDefaultLevels() {
        if (levelRepository.count() > 0) {
            // Levels already exist
            return;
        }
        
        // Create the 4 evaluation levels
        List<IloLevel> levels = new ArrayList<>();
        
        IloLevel veryStrong = new IloLevel();
        veryStrong.setMinPercent(80);
        veryStrong.setMaxPercent(100);
        veryStrong.setLevelName("Rất mạnh");
        veryStrong.setSuggestion("Bạn có năng lực rất nổi trội trong lĩnh vực này. Đây là thế mạnh lớn nhất của bạn.");
        levels.add(veryStrong);
        
        IloLevel strong = new IloLevel();
        strong.setMinPercent(60);
        strong.setMaxPercent(79);
        strong.setLevelName("Mạnh");
        strong.setSuggestion("Bạn có năng lực tốt trong lĩnh vực này. Đây là một thế mạnh của bạn.");
        levels.add(strong);
        
        IloLevel average = new IloLevel();
        average.setMinPercent(40);
        average.setMaxPercent(59);
        average.setLevelName("Trung bình");
        average.setSuggestion("Bạn có năng lực trung bình trong lĩnh vực này. Bạn có thể phát triển thêm.");
        levels.add(average);
        
        IloLevel weak = new IloLevel();
        weak.setMinPercent(0);
        weak.setMaxPercent(39);
        weak.setLevelName("Yếu");
        weak.setSuggestion("Đây không phải là thế mạnh của bạn, nhưng bạn vẫn có thể cải thiện nếu muốn.");
        levels.add(weak);
        
        levelRepository.saveAll(levels);
    }

    /**
     * Initialize default career mappings if they don't exist
     */
    @Transactional
    public void initializeDefaultCareerMappings() {
        if (careerMapRepository.count() > 0) {
            // Career mappings already exist
            return;
        }
        
        // Get all domains first
        Map<String, IloDomain> domainMap = getAllDomains().stream()
                .collect(Collectors.toMap(IloDomain::getCode, domain -> domain));
        
        List<IloCareerMap> careerMappings = new ArrayList<>();
        
        // LANG domain careers
        addCareerMapping(careerMappings, domainMap.get("LANG"), "Biên tập viên", "Editor", 100);
        addCareerMapping(careerMappings, domainMap.get("LANG"), "Phiên dịch viên", "Translator/Interpreter", 95);
        addCareerMapping(careerMappings, domainMap.get("LANG"), "Giáo viên ngoại ngữ", "Language Teacher", 90);
        addCareerMapping(careerMappings, domainMap.get("LANG"), "Nhà báo", "Journalist", 85);
        addCareerMapping(careerMappings, domainMap.get("LANG"), "PR - Truyền thông", "PR Specialist", 80);
        addCareerMapping(careerMappings, domainMap.get("LANG"), "Content Marketing", "Content Marketer", 75);
        
        // LOGIC domain careers
        addCareerMapping(careerMappings, domainMap.get("LOGIC"), "Lập trình viên", "Software Developer", 100);
        addCareerMapping(careerMappings, domainMap.get("LOGIC"), "Kỹ sư dữ liệu", "Data Engineer", 95);
        addCareerMapping(careerMappings, domainMap.get("LOGIC"), "Phân tích kinh doanh", "Business Analyst", 90);
        addCareerMapping(careerMappings, domainMap.get("LOGIC"), "Phân tích tài chính", "Financial Analyst", 85);
        addCareerMapping(careerMappings, domainMap.get("LOGIC"), "Nghiên cứu viên", "Researcher", 80);
        addCareerMapping(careerMappings, domainMap.get("LOGIC"), "Kế toán viên", "Accountant", 75);
        
        // DESIGN domain careers
        addCareerMapping(careerMappings, domainMap.get("DESIGN"), "Thiết kế đồ họa", "Graphic Designer", 100);
        addCareerMapping(careerMappings, domainMap.get("DESIGN"), "Thiết kế thời trang", "Fashion Designer", 95);
        addCareerMapping(careerMappings, domainMap.get("DESIGN"), "Kiến trúc sư", "Architect", 90);
        addCareerMapping(careerMappings, domainMap.get("DESIGN"), "Thiết kế nội thất", "Interior Designer", 85);
        addCareerMapping(careerMappings, domainMap.get("DESIGN"), "Thiết kế UX/UI", "UX/UI Designer", 80);
        addCareerMapping(careerMappings, domainMap.get("DESIGN"), "Hoạ sỹ", "Artist", 75);
        
        // PEOPLE domain careers
        addCareerMapping(careerMappings, domainMap.get("PEOPLE"), "Quản lý nhân sự", "HR Manager", 100);
        addCareerMapping(careerMappings, domainMap.get("PEOPLE"), "Chuyên viên tuyển dụng", "Recruiter", 95);
        addCareerMapping(careerMappings, domainMap.get("PEOPLE"), "Nhân viên xã hội", "Social Worker", 90);
        addCareerMapping(careerMappings, domainMap.get("PEOPLE"), "Cố vấn tâm lý", "Counselor/Therapist", 85);
        addCareerMapping(careerMappings, domainMap.get("PEOPLE"), "Giáo viên", "Teacher", 80);
        addCareerMapping(careerMappings, domainMap.get("PEOPLE"), "Quản lý dự án", "Project Manager", 75);
        
        // MECH domain careers
        addCareerMapping(careerMappings, domainMap.get("MECH"), "Kỹ sư cơ khí", "Mechanical Engineer", 100);
        addCareerMapping(careerMappings, domainMap.get("MECH"), "Kỹ thuật viên điện tử", "Electronics Technician", 95);
        addCareerMapping(careerMappings, domainMap.get("MECH"), "Huấn luyện viên thể thao", "Sports Coach", 90);
        addCareerMapping(careerMappings, domainMap.get("MECH"), "Vận động viên chuyên nghiệp", "Professional Athlete", 85);
        addCareerMapping(careerMappings, domainMap.get("MECH"), "Nông nghiệp", "Agricultural Worker", 80);
        addCareerMapping(careerMappings, domainMap.get("MECH"), "Y tá", "Nurse", 75);
        
        // Save all mappings
        careerMapRepository.saveAll(careerMappings);
    }
    
    /**
     * Helper method to add a career mapping to the list
     */
    private void addCareerMapping(List<IloCareerMap> mappings, IloDomain domain, String careerField, String description, Integer priority) {
        IloCareerMap mapping = new IloCareerMap();
        mapping.setDomain(domain);
        mapping.setCareerField(careerField);
        mapping.setDescription(description);
        mapping.setPriority(priority);
        mappings.add(mapping);
    }
}
