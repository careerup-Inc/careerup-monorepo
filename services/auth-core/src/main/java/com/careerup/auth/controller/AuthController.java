package com.careerup.auth.controller;

import com.careerup.auth.model.User;
import com.careerup.auth.service.AuthService;
import io.swagger.annotations.Api;
import io.swagger.annotations.ApiOperation;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/v1/auth")
@RequiredArgsConstructor
@Api(tags = "Authentication")
public class AuthController {
    private final AuthService authService;

    @PostMapping("/register")
    @ApiOperation("Register a new user")
    public ResponseEntity<User> register(
            @RequestParam String email,
            @RequestParam String password,
            @RequestParam String firstName,
            @RequestParam String lastName
    ) {
        User user = authService.register(email, password, firstName, lastName);
        return ResponseEntity.ok(user);
    }

    @PostMapping("/login")
    @ApiOperation("Login user")
    public ResponseEntity<Map<String, String>> login(
            @RequestParam String email,
            @RequestParam String password
    ) {
        Map<String, String> tokens = authService.login(email, password);
        return ResponseEntity.ok(tokens);
    }

    @PostMapping("/refresh")
    @ApiOperation("Refresh access token")
    public ResponseEntity<Map<String, String>> refreshToken(
            @RequestParam String refreshToken
    ) {
        Map<String, String> tokens = authService.refreshToken(refreshToken);
        return ResponseEntity.ok(tokens);
    }
} 