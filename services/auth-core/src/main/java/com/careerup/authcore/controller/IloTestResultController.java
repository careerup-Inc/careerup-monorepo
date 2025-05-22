package com.careerup.authcore.controller;

import com.careerup.authcore.model.IloTestResult;
import com.careerup.authcore.service.IloTestResultService;
import com.careerup.authcore.service.AuthService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.Optional;

@RestController
@RequestMapping("/api/v1/ilo")
@RequiredArgsConstructor
public class IloTestResultController {
    private final IloTestResultService iloTestResultService;
    private final AuthService authService;

    @PostMapping("/result")
    public ResponseEntity<IloTestResult> submitIloTestResult(
            @RequestBody Map<String, Object> payload,
            HttpServletRequest request
    ) {
        String authHeader = request.getHeader("Authorization");
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        String token = authHeader.substring(7);
        String userId = authService.validateToken(token).getId().toString();
        String resultData = payload.get("resultData").toString();
        IloTestResult result = iloTestResultService.saveResult(UUID.fromString(userId), resultData);
        return ResponseEntity.status(HttpStatus.CREATED).body(result);
    }

    @GetMapping("/results")
    public ResponseEntity<List<IloTestResult>> getIloTestResults(HttpServletRequest request) {
        String authHeader = request.getHeader("Authorization");
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        String token = authHeader.substring(7);
        String userId = authService.validateToken(token).getId().toString();
        List<IloTestResult> results = iloTestResultService.getResultsByUserId(UUID.fromString(userId));
        return ResponseEntity.ok(results);
    }

    @GetMapping("/result/{id}")
    public ResponseEntity<?> getIloTestResultById(
            @PathVariable("id") String idStr,
            HttpServletRequest request
    ) {
        String authHeader = request.getHeader("Authorization");
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }

        String token = authHeader.substring(7);
        String userId = authService.validateToken(token).getId().toString();

        // Convert string ID to Long, handling potential NumberFormatException
        Long id;
        try {
            id = Long.parseLong(idStr);
        } catch (NumberFormatException e) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Invalid ID format: " + idStr));
        }

        Optional<IloTestResult> resultOpt = iloTestResultService.getResultById(id);
        if (resultOpt.isEmpty()) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Result not found"));
        }

        IloTestResult result = resultOpt.get();
        if (!result.getUserId().toString().equals(userId)) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "You don't have permission to access this result"));
        }

        return ResponseEntity.ok(result);
    }
}
