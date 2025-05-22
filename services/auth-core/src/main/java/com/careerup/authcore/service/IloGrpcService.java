package com.careerup.authcore.service;

import com.careerup.authcore.model.*;
import com.careerup.authcore.repository.IloDomainScoreRepository;
import com.careerup.authcore.repository.IloTestResultRepository;
import com.careerup.proto.v1.*;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class IloGrpcService extends IloServiceGrpc.IloServiceImplBase {
    private final IloTestResultService iloTestResultService;
    private final IloQuestionService iloQuestionService;
    private final IloDomainService iloDomainService;
    private final IloTestResultRepository iloTestResultRepository;
    private final IloDomainScoreRepository iloDomainScoreRepository;

    @Override
    public void submitIloTestResult(SubmitIloTestResultRequest request,
            StreamObserver<SubmitIloTestResultResponse> responseObserver) {
        // Use the new method with structured answers
        com.careerup.authcore.model.IloTestResult saved = iloTestResultService.saveResultWithAnswers(
                UUID.fromString(request.getUserId()),
                request.getRawResultData(),
                request.getAnswersList());

        // Build the response with all fields including domain scores
        com.careerup.proto.v1.IloTestResult.Builder resultBuilder = com.careerup.proto.v1.IloTestResult.newBuilder()
                .setId(saved.getId().toString())
                .setUserId(saved.getUserId().toString())
                .setResultData(saved.getResultData())
                .setCreatedAt(saved.getCreatedAt().toString());

        // Add domain scores
        for (com.careerup.authcore.model.IloDomainScore score : saved.getDomainScores()) {
            com.careerup.proto.v1.IloDomainScore protoScore = com.careerup.proto.v1.IloDomainScore.newBuilder()
                    .setDomainCode(score.getDomain().getCode())
                    .setRawScore(score.getRawScore())
                    .setPercent(score.getPercentScore())
                    .setLevel(score.getLevel().getLevelName())
                    .setRank(score.getRank())
                    .build();

            resultBuilder.addScores(protoScore);
        }

        // Add top domains
        List<String> topDomains = saved.getDomainScores().stream()
                .sorted((a, b) -> Float.compare(b.getPercentScore(), a.getPercentScore()))
                .limit(3)
                .map(score -> score.getDomain().getCode())
                .collect(Collectors.toList());

        resultBuilder.addAllTopDomains(topDomains);

        // Add suggested careers
        if (saved.getSuggestedCareers() != null && !saved.getSuggestedCareers().isEmpty()) {
            String[] careers = saved.getSuggestedCareers().split(",");
            resultBuilder.addAllSuggestedCareers(List.of(careers));
        }

        SubmitIloTestResultResponse response = SubmitIloTestResultResponse.newBuilder()
                .setResult(resultBuilder.build())
                .build();

        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getIloTestResult(GetIloTestResultRequest request,
            StreamObserver<GetIloTestResultResponse> responseObserver) {
        try {
            String resultId = request.getResultId();
            // Fetch result by ID WITH domain scores
            com.careerup.authcore.model.IloTestResult result = iloTestResultRepository.findByIdWithDomainScores(Long.parseLong(resultId));

            if (result == null) {
                responseObserver.onError(
                        io.grpc.Status.NOT_FOUND.withDescription("Test result not found for result ID: " + resultId)
                                .asRuntimeException());
                return;
            }

            // Explicitly fetch domain scores
            List<com.careerup.authcore.model.IloDomainScore> domainScores = result.getDomainScores();
            System.out.println("Found " + (domainScores != null ? domainScores.size() : 0) +
                                " domain scores for result ID " + resultId);

            GetIloTestResultResponse response = GetIloTestResultResponse.newBuilder()
                    .setResult(buildTestResultProto(result).build())
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception e) {
            System.err.println("Error in getIloTestResult: " + e.getMessage());
            e.printStackTrace();
            responseObserver.onError(
                    io.grpc.Status.INTERNAL
                            .withDescription("Error retrieving ILO test result: " + e.getMessage())
                            .asRuntimeException());
        }
    }

    @Override
    @Transactional(readOnly = true)
    public void getIloTestResults(GetIloTestResultsRequest request,
            StreamObserver<GetIloTestResultsResponse> responseObserver) {
        try {
            // Safely parse userId
            UUID userId;
            try {
                userId = UUID.fromString(request.getUserId());
            } catch (IllegalArgumentException e) {
                responseObserver.onError(
                        io.grpc.Status.INVALID_ARGUMENT
                                .withDescription("Invalid user ID format: " + request.getUserId())
                                .asRuntimeException());
                return;
            }

            // Try cache first - if available, use it to fetch results by ID
            Object cached = iloDomainService.getCachedUserResult(request.getUserId());
            if (cached != null && cached instanceof List<?>) {
                try {
                    @SuppressWarnings("unchecked")
                    List<String> cachedIds = (List<String>) cached;
                    if (!cachedIds.isEmpty()) {
                        // Build response from cached IDs
                        GetIloTestResultsResponse.Builder respBuilder = GetIloTestResultsResponse.newBuilder();

                        for (String idStr : cachedIds) {
                            try {
                                Long id = Long.parseLong(idStr);
                                iloTestResultRepository.findById(id)
                                        .ifPresent(r -> respBuilder.addResults(buildTestResultProto(r).build()));
                            } catch (Exception e) {
                                System.err.println("Error processing cached ID: " + idStr + " - " + e.getMessage());
                            }
                        }

                        GetIloTestResultsResponse response = respBuilder.build();
                        responseObserver.onNext(response);
                        responseObserver.onCompleted();
                        return;
                    }
                } catch (Exception e) {
                    System.err.println("Error processing cached results: " + e.getMessage());
                    // Continue to fetch from database
                }
            }

            // No valid cache, fetch from database
            List<com.careerup.authcore.model.IloTestResult> results = iloTestResultRepository
                    .findByUserIdOrderByCreatedAtDesc(userId);

            GetIloTestResultsResponse.Builder respBuilder = GetIloTestResultsResponse.newBuilder();

            if (results != null && !results.isEmpty()) {
                for (com.careerup.authcore.model.IloTestResult result : results) {
                    try {
                        respBuilder.addResults(buildTestResultProto(result).build());
                    } catch (Exception e) {
                        System.err.println("Error processing result " + result.getId() + ": " + e.getMessage());
                    }
                }
            }

            // Save only the IDs to cache to avoid serialization issues
            GetIloTestResultsResponse response = respBuilder.build();
            try {
                List<String> resultIds = response.getResultsList().stream()
                        .map(com.careerup.proto.v1.IloTestResult::getId)
                        .collect(Collectors.toList());

                // Only cache if we found results
                if (!resultIds.isEmpty()) {
                    iloDomainService.cacheUserResult(request.getUserId(), resultIds);
                }
            } catch (Exception e) {
                System.err.println("Error caching results: " + e.getMessage());
            }

            responseObserver.onNext(response);
            responseObserver.onCompleted();

        } catch (Exception e) {
            System.err.println("Error in getIloTestResults: " + e.getMessage());
            e.printStackTrace();
            responseObserver.onError(
                    io.grpc.Status.INTERNAL
                            .withDescription("Error retrieving ILO test results: " + e.getMessage())
                            .asRuntimeException());
        }
    }

    @Override
    @Transactional(readOnly = true)
    public void getIloTest(GetIloTestRequest request, StreamObserver<GetIloTestResponse> responseObserver) {
        com.careerup.proto.v1.GetIloTestResponse.Builder respBuilder = com.careerup.proto.v1.GetIloTestResponse
                .newBuilder();

        // Fetch questions from database
        List<IloQuestion> questions = iloQuestionService.getAllQuestions();
        System.out.println("Retrieved " + questions.size() + " questions from database");

        // Debug: Print domain distribution
        Map<String, Integer> domainCounts = new HashMap<>();
        for (IloQuestion question : questions) {
            if (!question.getQuestionDomains().isEmpty()) {
                String domainCode = question.getQuestionDomains().get(0).getDomain().getCode();
                domainCounts.put(domainCode, domainCounts.getOrDefault(domainCode, 0) + 1);
            }
        }
        System.out.println("Domain distribution: " + domainCounts);

        // Add questions with all fields
        int questionsWithDomains = 0;
        int questionsWithoutDomains = 0;

        for (IloQuestion question : questions) {
            com.careerup.proto.v1.IloTestQuestion.Builder questionBuilder = com.careerup.proto.v1.IloTestQuestion
                    .newBuilder()
                    .setId(question.getId().toString())
                    .setQuestionNumber(question.getQuestionNumber())
                    .setText(question.getQuestionText());

            // Add domain code if available
            if (!question.getQuestionDomains().isEmpty()) {
                String domainCode = question.getQuestionDomains().get(0).getDomain().getCode();
                questionBuilder.setDomainCode(domainCode);
                questionsWithDomains++;
            } else {
                questionsWithoutDomains++;
                // Log questions without domains
                System.out.println("Question without domain: #" + question.getQuestionNumber() + " - "
                        + question.getQuestionText().substring(0, Math.min(50, question.getQuestionText().length()))
                        + "...");
            }

            // Add options
            for (String option : question.getOptions()) {
                questionBuilder.addOptions(option);
            }

            respBuilder.addQuestions(questionBuilder.build());
        }

        // Log domain assignment stats
        System.out.println("Questions with domain assignments: " + questionsWithDomains);
        System.out.println("Questions WITHOUT domain assignments: " + questionsWithoutDomains);

        // Add domains
        java.util.List<com.careerup.authcore.model.IloDomain> domains = iloDomainService.getAllDomains();
        for (com.careerup.authcore.model.IloDomain domain : domains) {
            com.careerup.proto.v1.IloDomain protoDomain = com.careerup.proto.v1.IloDomain.newBuilder()
                    .setCode(domain.getCode())
                    .setName(domain.getName())
                    .setDescription(domain.getDescription() != null ? domain.getDescription() : "")
                    .build();

            respBuilder.addDomains(protoDomain);
        }

        // Add levels
        java.util.List<com.careerup.authcore.model.IloLevel> levels = iloDomainService.getAllLevels();
        for (com.careerup.authcore.model.IloLevel level : levels) {
            com.careerup.proto.v1.IloLevel protoLevel = com.careerup.proto.v1.IloLevel.newBuilder()
                    .setMinPercent(level.getMinPercent())
                    .setMaxPercent(level.getMaxPercent())
                    .setLevelName(level.getLevelName())
                    .setSuggestion(level.getSuggestion() != null ? level.getSuggestion() : "")
                    .build();

            respBuilder.addLevels(protoLevel);
        }

        responseObserver.onNext(respBuilder.build());
        responseObserver.onCompleted();
    }

    @Override
    @Transactional(readOnly = true)
    public void getIloCareerSuggestions(com.careerup.proto.v1.GetIloCareerSuggestionsRequest request,
            io.grpc.stub.StreamObserver<com.careerup.proto.v1.GetIloCareerSuggestionsResponse> responseObserver) {
        java.util.List<String> domainCodes = request.getDomainCodesList();
        int limit = request.getLimit() > 0 ? request.getLimit() : 5; // Default to 5 if not specified

        // Convert domain codes to domain scores (using dummy scores for search only)
        java.util.List<com.careerup.authcore.model.IloDomain> allDomains = iloDomainService.getAllDomains();
        java.util.List<com.careerup.authcore.model.IloDomainScore> dummyScores = domainCodes.stream()
                .map(code -> {
                    com.careerup.authcore.model.IloDomain domain = allDomains.stream()
                            .filter(d -> d.getCode().equals(code))
                            .findFirst()
                            .orElse(null);
                    if (domain == null) {
                        return null;
                    }
                    com.careerup.authcore.model.IloDomainScore score = new com.careerup.authcore.model.IloDomainScore();
                    score.setDomain(domain);
                    score.setPercentScore(100.0f); // Use max score for search
                    return score;
                })
                .filter(score -> score != null)
                .collect(java.util.stream.Collectors.toList());

        // Get career suggestions
        java.util.List<com.careerup.authcore.model.IloCareerMap> careers = iloDomainService
                .getCareerSuggestions(dummyScores, limit);

        // Build response
        com.careerup.proto.v1.GetIloCareerSuggestionsResponse.Builder respBuilder = com.careerup.proto.v1.GetIloCareerSuggestionsResponse
                .newBuilder();
        for (com.careerup.authcore.model.IloCareerMap career : careers) {
            com.careerup.proto.v1.IloCareerSuggestion suggestion = com.careerup.proto.v1.IloCareerSuggestion
                    .newBuilder()
                    .setDomainCode(career.getDomain().getCode())
                    .setCareerField(career.getCareerField())
                    .build();

            respBuilder.addSuggestions(suggestion);
        }

        responseObserver.onNext(respBuilder.build());
        responseObserver.onCompleted();
    }

    /**
     * Helper method to build a protobuf IloTestResult from a domain entity
     */
    private com.careerup.proto.v1.IloTestResult.Builder buildTestResultProto(
            com.careerup.authcore.model.IloTestResult r) {
        com.careerup.proto.v1.IloTestResult.Builder resultBuilder = com.careerup.proto.v1.IloTestResult.newBuilder()
                .setId(r.getId().toString())
                .setUserId(r.getUserId().toString())
                .setResultData(r.getResultData())
                .setCreatedAt(r.getCreatedAt().toString());

        // Get domain scores directly from repository to avoid
        // LazyInitializationException
        List<com.careerup.authcore.model.IloDomainScore> domainScores = iloDomainScoreRepository
                .findByTestResultIdOrderByRankAsc(r.getId());

        if (domainScores != null && !domainScores.isEmpty()) {
            for (com.careerup.authcore.model.IloDomainScore score : domainScores) {
                try {
                    com.careerup.proto.v1.IloDomainScore protoScore = com.careerup.proto.v1.IloDomainScore.newBuilder()
                            .setDomainCode(score.getDomain().getCode())
                            .setRawScore(score.getRawScore())
                            .setPercent(score.getPercentScore())
                            .setLevel(score.getLevel().getLevelName())
                            .setRank(score.getRank())
                            .build();
                    resultBuilder.addScores(protoScore);
                } catch (Exception e) {
                    System.err.println("Error processing domain score: " + e.getMessage());
                }
            }

            java.util.List<String> topDomains = domainScores.stream()
                    .sorted((a, b) -> Float.compare(b.getPercentScore(), a.getPercentScore()))
                    .limit(3)
                    .map(score -> score.getDomain().getCode())
                    .collect(java.util.stream.Collectors.toList());
            resultBuilder.addAllTopDomains(topDomains);
        }

        if (r.getSuggestedCareers() != null && !r.getSuggestedCareers().isEmpty()) {
            String[] careers = r.getSuggestedCareers().split(",");
            resultBuilder.addAllSuggestedCareers(java.util.List.of(careers));
        }

        return resultBuilder;
    }
}
