package com.careerup.authcore.service;

import com.careerup.authcore.model.User;
import com.careerup.authcore.model.dto.TokenResponse;
import com.careerup.authcore.repository.UserRepository;
import com.careerup.authcore.security.UserDetailsImpl;
import com.careerup.proto.v1.LoginRequest;
import com.careerup.proto.v1.LoginResponse;
import com.careerup.proto.v1.RegisterRequest;
import com.careerup.proto.v1.RegisterResponse;
import com.careerup.proto.v1.ValidateTokenRequest;
import com.careerup.proto.v1.ValidateTokenResponse;
import com.careerup.proto.v1.RefreshTokenRequest;
import com.careerup.proto.v1.RefreshTokenResponse;
import com.careerup.proto.v1.UpdateUserRequest;
import com.careerup.proto.v1.UpdateUserResponse;

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
        try {
            String email = jwtService.extractUsername(token);
            if (email == null) {
                throw new RuntimeException("Invalid token");
            }
            
            User user = userRepository.findByEmail(email)
                    .orElseThrow(() -> new RuntimeException("User not found"));
            
            // Convert to UserDetails for validation
            UserDetails userDetails = UserDetailsImpl.build(user);
            
            if (!jwtService.isTokenValid(token, userDetails)) {
                throw new RuntimeException("Token is not valid");
            }
            
            return user;
        } catch (Exception e) {
            throw new RuntimeException("Error validating token", e);
        }
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

    // grpc methods
    public LoginResponse grpcLogin(LoginRequest request) {
        String email = request.getEmail();
        String password = request.getPassword();
        TokenResponse tokenResponse = login(email, password);
        
        return LoginResponse.newBuilder()
            .setAccessToken(tokenResponse.accessToken())
            .setRefreshToken(tokenResponse.refreshToken())
            .setExpireIn(tokenResponse.expiresIn())
            .build();
    }

    public RegisterResponse grpcRegister(RegisterRequest request) {
        User user = register(request.getEmail(), request.getPassword(), request.getFirstName(), request.getLastName());

        com.careerup.proto.v1.User protoUser = com.careerup.proto.v1.User.newBuilder()
        .setId(user.getId() != null ? user.getId().toString() : "")
        .setEmail(user.getEmail())
        .setFirstName(user.getFirstName() != null ? user.getFirstName() : "")
        .setLastName(user.getLastName() != null ? user.getLastName() : "")
        .setHometown(user.getHometown() != null ? user.getHometown() : "")
        .build();
        
        return RegisterResponse.newBuilder()
            .setUser(protoUser)
            .build();
    }

    public ValidateTokenResponse grpcValidateToken(ValidateTokenRequest request) {
        String token = request.getToken();
        User user = validateToken(token);
        
        com.careerup.proto.v1.User protoUser = com.careerup.proto.v1.User.newBuilder()
            .setId(user.getId() != null ? user.getId().toString() : "")
            .setEmail(user.getEmail())
            .setFirstName(user.getFirstName() != null ? user.getFirstName() : "")
            .setLastName(user.getLastName() != null ? user.getLastName() : "")
            .setHometown(user.getHometown() != null ? user.getHometown() : "")
            .build();
        
        return ValidateTokenResponse.newBuilder()
            .setUser(protoUser)
            .build();
    }

    public RefreshTokenResponse grpcRefreshToken(RefreshTokenRequest request) {
        String refreshToken = request.getRefreshToken();
        TokenResponse tokenResponse = refreshToken(refreshToken);
        
        return RefreshTokenResponse.newBuilder()
            .setAccessToken(tokenResponse.accessToken())
            .setRefreshToken(tokenResponse.refreshToken())
            .setExpireIn(tokenResponse.expiresIn())
            .build();
    }

    public UpdateUserResponse grpcUpdateUser(UpdateUserRequest request) {
        String token = request.getToken();
        // find user to be update by token
        User _user = validateToken(token);
        String firstName = request.getFirstName();
        String lastName = request.getLastName();
        String hometown = request.getHometown();
        List<String> interests = request.getInterestsList();

        User user = updateUser(_user.getEmail(), firstName, lastName, hometown, interests);

        com.careerup.proto.v1.User protoUser = com.careerup.proto.v1.User.newBuilder()
            .setFirstName(user.getFirstName() != null ? user.getFirstName() : "")
            .setLastName(user.getLastName() != null ? user.getLastName() : "")
            .setHometown(user.getHometown() != null ? user.getHometown() : "")
            .build();

        return UpdateUserResponse.newBuilder()
            .setUser(protoUser)
            .build();
    }
}