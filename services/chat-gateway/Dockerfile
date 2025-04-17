FROM eclipse-temurin:21.0.6_7-jre-jammy

WORKDIR /app

COPY build/libs/*.jar app.jar

EXPOSE 8082

USER nobody

ENTRYPOINT ["java", "-jar", "app.jar"] 