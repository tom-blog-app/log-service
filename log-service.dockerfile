FROM alpine:latest

RUN mkdir /app

COPY log-service /app

CMD [ "/app/log-service"]