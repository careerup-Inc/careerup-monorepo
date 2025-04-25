package com.careerup.auth.service;

import com.careerup.auth.model.User;
import com.careerup.auth.repository.UserRepository;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import io.jsonwebtoken.security.Keys;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.security.Key;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class AuthService {
    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final Key jwtSecretKey = Keys.secretKeyFor(SignatureAlgorithm.HS256);

    public User register(String email, String password, String firstName, String lastName) {
        if (userRepository.existsByEmail(email)) {
            throw new RuntimeException("Email already exists");
        }

        User user = User.builder()
                .email(email)
                .password(passwordEncoder.encode(password))
                .firstName(firstName)
                .lastName(lastName)
                .isActive(true)
                .build();

        return userRepository.save(user);
    }

    public Map<String, String> login(String email, String password) {
        User user = userRepository.findByEmail(email)
                .orElseThrow(() -> new RuntimeException("Invalid credentials"));

        if (!passwordEncoder.matches(password, user.getPassword())) {
            throw new RuntimeException("Invalid credentials");
        }

        if (!user.isActive()) {
            throw new RuntimeException("Account is not active");
        }

        return generateTokens(user);
    }

    public Map<String, String> refreshToken(String refreshToken) {
        try {
            String userId = Jwts.parserBuilder()
                    .setSigningKey(jwtSecretKey)
                    .build()
                    .parseClaimsJws(refreshToken)
                    .getBody()
                    .getSubject();

            User user = userRepository.findById(UUID.fromString(userId))
                    .orElseThrow(() -> new RuntimeException("User not found"));

            if (!user.isActive()) {
                throw new RuntimeException("Account is not active");
            }

            return generateTokens(user);
        } catch (Exception e) {
            throw new RuntimeException("Invalid refresh token");
        }
    }

    private Map<String, String> generateTokens(User user) {
        Date now = new Date();
        Date accessTokenExpiry = new Date(now.getTime() + 15 * 60 * 1000); // 15 minutes
        Date refreshTokenExpiry = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000); // 7 days

        String accessToken = Jwts.builder()
                .setSubject(user.getId().toString())
                .setIssuedAt(now)
                .setExpiration(accessTokenExpiry)
                .signWith(jwtSecretKey)
                .compact();

        String refreshToken = Jwts.builder()
                .setSubject(user.getId().toString())
                .setIssuedAt(now)
                .setExpiration(refreshTokenExpiry)
                .signWith(jwtSecretKey)
                .compact();

        Map<String, String> tokens = new HashMap<>();
        tokens.put("access_token", accessToken);
        tokens.put("refresh_token", refreshToken);
        tokens.put("expires_in", String.valueOf(accessTokenExpiry.getTime()));

        return tokens;
    }
} 