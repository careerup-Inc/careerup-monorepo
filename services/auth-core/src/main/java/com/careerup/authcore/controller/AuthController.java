package com.careerup.authcore.controller;

import com.careerup.authcore.model.User;
import com.careerup.authcore.model.dto.LoginRequest;
import com.careerup.authcore.model.dto.RefreshTokenRequest;
import com.careerup.authcore.model.dto.RegisterRequest;
import com.careerup.authcore.model.dto.TokenResponse;
import com.careerup.authcore.model.dto.UpdateUserRequest;
import com.careerup.authcore.service.AuthService;

import jakarta.servlet.http.HttpServletRequest;
import lombok.RequiredArgsConstructor;
import com.careerup.authcore.model.dto.UserDTO;

import java.util.HashMap;
import java.util.Map;
import java.util.Collections;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/auth")
@RequiredArgsConstructor
public class AuthController {
    private final AuthService authService;

    @PostMapping("/register")
    public ResponseEntity<User> register(
        @RequestBody RegisterRequest request
    ) {
        User user = authService.register(
            request.getEmail(),
            request.getPassword(),
            request.getFirstName(),
            request.getLastName()
        );
        return ResponseEntity.status(HttpStatus.CREATED).body(user);
    }

    @PostMapping("/login")
    public ResponseEntity<Map<String, Object>> login(
        @RequestBody LoginRequest request
    ) {
        try {
            TokenResponse tokenResponse = authService.login(request.getEmail(), request.getPassword());

            Map<String, Object> response = new HashMap<>();
            response.put("access_token", tokenResponse.accessToken());
            response.put("refresh_token", tokenResponse.refreshToken()); 
            response.put("expires_in", tokenResponse.expiresIn());
            
            return ResponseEntity.ok(response);
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(null);
        }
    }

    @PostMapping("/refresh")
    public ResponseEntity<Map<String, Object>> refreshToken(@RequestBody RefreshTokenRequest request) {
        try {
            TokenResponse tokenResponse = authService.refreshToken(request.getRefreshToken());
            
            Map<String, Object> response = new HashMap<>();
            response.put("access_token", tokenResponse.accessToken());
            response.put("refresh_token", tokenResponse.refreshToken());
            response.put("expires_in", tokenResponse.expiresIn());
            
            return ResponseEntity.ok(response);
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(
                Collections.singletonMap("error", "Invalid refresh token")
            );
        }
    }

    @GetMapping("/validate")
    public ResponseEntity<UserDTO> validateToken(HttpServletRequest request) {
        // Extract token from Authorization header
        String authHeader = request.getHeader("Authorization");
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        
        String token = authHeader.substring(7);
        
        try {
            // Call the service to validate the token
            User user = authService.validateToken(token);
            
            // Convert to DTO for response
            UserDTO userDTO = new UserDTO();
            userDTO.setId(user.getId());
            userDTO.setEmail(user.getEmail());
            userDTO.setFirstName(user.getFirstName());
            userDTO.setLastName(user.getLastName());
            userDTO.setHometown(user.getHometown());
            userDTO.setInterests(user.getInterests());
            userDTO.setIsActive(true);
            
            return ResponseEntity.ok(userDTO);
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
    }

    @GetMapping("/me")
    public ResponseEntity<User> getCurrentUser(@RequestParam("email") String email) {
        User user = authService.getCurrentUser(email);
        return ResponseEntity.ok(user);
    }

    @PutMapping("/me")
    public ResponseEntity<User> updateUser(
        @RequestBody UpdateUserRequest request
    ) {
        User user = authService.updateUser(
            request.getEmail(),
            request.getFirstName(),
            request.getLastName(),
            request.getHometown(),
            request.getInterests()
        );
        return ResponseEntity.ok(user);
    }
} 