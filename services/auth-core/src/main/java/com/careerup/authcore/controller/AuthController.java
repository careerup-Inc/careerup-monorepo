package com.careerup.authcore.controller;

import com.careerup.authcore.model.User;
import com.careerup.authcore.service.AuthService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/v1/auth")
@RequiredArgsConstructor
public class AuthController {
    private final AuthService authService;

    @PostMapping("/register")
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
    public ResponseEntity<String> login(
        @RequestParam String email,
        @RequestParam String password
    ) {
        String token = authService.login(email, password);
        return ResponseEntity.ok(token);
    }

    @PostMapping("/validate")
    public ResponseEntity<User> validateToken(@RequestParam String token) {
        User user = authService.validateToken(token);
        return ResponseEntity.ok(user);
    }

    @GetMapping("/me")
    public ResponseEntity<User> getCurrentUser(@RequestParam String email) {
        User user = authService.getCurrentUser(email);
        return ResponseEntity.ok(user);
    }

    @PutMapping("/me")
    public ResponseEntity<User> updateUser(
        @RequestParam String email,
        @RequestParam(required = false) String firstName,
        @RequestParam(required = false) String lastName,
        @RequestParam(required = false) String hometown,
        @RequestParam(required = false) List<String> interests
    ) {
        User user = authService.updateUser(email, firstName, lastName, hometown, interests);
        return ResponseEntity.ok(user);
    }
} 