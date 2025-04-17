FROM eclipse-temurin:21.0.6_7-jre-jammy

WORKDIR /app

COPY . .

EXPOSE 8081

USER nobody

CMD ["./gradlew", "bootRun"]