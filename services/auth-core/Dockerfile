FROM eclipse-temurin:21.0.6_7-jre-jammy

WORKDIR /app

COPY build/libs/*.jar app.jar

# HTTP
EXPOSE 8081
# gRPC
EXPOSE 9091

USER nobody

ENTRYPOINT ["java", "-jar", "app.jar"] 