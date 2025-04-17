FROM quay.io/quarkus/quarkus-micro-image:2.0

WORKDIR /work/
COPY target/*-runner /work/application

EXPOSE 8083

USER nobody

CMD ["./application", "-Dquarkus.http.host=0.0.0.0"] 