package com.careerup.authcore.service;

import com.careerup.authcore.model.User;
import com.careerup.authcore.model.dto.TokenResponse;
import com.careerup.authcore.repository.UserRepository;
import com.careerup.authcore.security.UserDetailsImpl;
import lombok.RequiredArgsConstructor;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@RequiredArgsConstructor
public class AuthService {
    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final JwtService jwtService;
    private final AuthenticationManager authenticationManager;

    @Transactional
    public User register(String email, String password, String firstName, String lastName) {
        if (userRepository.existsByEmail(email)) {
            throw new RuntimeException("Email already exists");
        }

        User user = new User();
        user.setEmail(email);
        user.setPassword(passwordEncoder.encode(password));
        user.setFirstName(firstName);
        user.setLastName(lastName);
        user.setInterests(List.of());

        return userRepository.save(user);
    }

    public TokenResponse login(String email, String password) {
        Authentication authentication = authenticationManager.authenticate(
            new UsernamePasswordAuthenticationToken(email, password)
        );
        SecurityContextHolder.getContext().setAuthentication(authentication);
        
        UserDetailsImpl userDetails = (UserDetailsImpl) authentication.getPrincipal();
        
        // Generate both tokens
        String accessToken = jwtService.generateToken(userDetails);
        String refreshToken = jwtService.generateRefreshToken(userDetails);
        
        // Get expiration time in seconds
        long expiresIn = jwtService.getExpiration() / 1000;
        
        return new TokenResponse(accessToken, refreshToken, expiresIn);
    }

    public TokenResponse refreshToken(String refreshToken) {
        try {
            // Extract username from the refresh token
            String username = jwtService.extractUsername(refreshToken);
            
            // Get the user details
            UserDetails userDetails = userRepository.findByEmail(username)
                .map(user -> new UserDetailsImpl(user))
                .orElseThrow(() -> new RuntimeException("User not found"));
            
            // Validate the refresh token
            if (!jwtService.isTokenValid(refreshToken, userDetails)) {
                throw new RuntimeException("Invalid refresh token");
            }
            
            // Generate new tokens
            String accessToken = jwtService.generateToken(userDetails);
            String newRefreshToken = jwtService.generateRefreshToken(userDetails);
            
            // Get expiration time in seconds
            long expiresIn = jwtService.getExpiration() / 1000;
            
            return new TokenResponse(accessToken, newRefreshToken, expiresIn);
        } catch (Exception e) {
            throw new RuntimeException("Error refreshing token", e);
        }
    }

    public User validateToken(String token) {
        String email = jwtService.extractUsername(token);
        return userRepository.findByEmail(email)
            .orElseThrow(() -> new RuntimeException("User not found"));
    }

    public User getCurrentUser(String email) {
        return userRepository.findByEmail(email)
            .orElseThrow(() -> new RuntimeException("User not found"));
    }

    @Transactional
    public User updateUser(String email, String firstName, String lastName, String hometown, List<String> interests) {
        User user = userRepository.findByEmail(email)
            .orElseThrow(() -> new RuntimeException("User not found"));

        if (firstName != null) user.setFirstName(firstName);
        if (lastName != null) user.setLastName(lastName);
        if (hometown != null) user.setHometown(hometown);
        if (interests != null) user.setInterests(interests);

        return userRepository.save(user);
    }
} 