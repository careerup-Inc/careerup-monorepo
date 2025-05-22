package com.careerup.authcore;

import com.careerup.authcore.service.AuthGrpcService;
import com.careerup.authcore.service.IloGrpcService;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.env.Environment;
import java.util.Arrays;

@Component
public class GrpcServerRunner implements CommandLineRunner {

    @Autowired
    private AuthGrpcService authGrpcService;

    @Autowired
    private IloGrpcService iloGrpcService;

    @Autowired
    private Environment environment;

    @Value("${grpc.enabled:true}")
    private boolean grpcEnabled;

    private Server server;

    @Override
    public void run(String... args) throws Exception {
        if (grpcEnabled && !Arrays.asList(environment.getActiveProfiles()).contains("test")) {
            int grpcPort = 9091;
            server = ServerBuilder.forPort(grpcPort)
            .addService(authGrpcService)
            .addService(iloGrpcService)
            .build()
            .start();
            System.out.println("gRPC server started on port " + grpcPort);
            Runtime.getRuntime().addShutdownHook(new Thread(() -> {
                if (server != null) server.shutdown();
            }));
            server.awaitTermination();
        }
    }
}