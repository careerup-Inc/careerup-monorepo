package com.careerup.authcore.service;

import com.careerup.proto.v1.AuthServiceGrpc;
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
import io.grpc.stub.StreamObserver;
import org.springframework.stereotype.Service;

@Service
public class AuthGrpcService extends AuthServiceGrpc.AuthServiceImplBase {

    private final AuthService authService;

    public AuthGrpcService(AuthService authService) {
        this.authService = authService;
    }

    @Override
    public void login(LoginRequest request, StreamObserver<LoginResponse> responseObserver) {
        LoginResponse response = authService.grpcLogin(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void register(RegisterRequest request, StreamObserver<RegisterResponse> responseObserver) {
        RegisterResponse response = authService.grpcRegister(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void validateToken(ValidateTokenRequest request, StreamObserver<ValidateTokenResponse> responseObserver) {
        ValidateTokenResponse response = authService.grpcValidateToken(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void refreshToken(RefreshTokenRequest request, StreamObserver<RefreshTokenResponse> responseObserver) {
        RefreshTokenResponse response = authService.grpcRefreshToken(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }

    @Override
    public void updateUser(UpdateUserRequest request, StreamObserver<UpdateUserResponse> responseObserver) {
        UpdateUserResponse response = authService.grpcUpdateUser(request);
        responseObserver.onNext(response);
        responseObserver.onCompleted();
    }
}
