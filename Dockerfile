#build stage
FROM golang:1.16.7-buster as build-env
ARG GH_TOKEN
RUN git config --global url."https://${GH_TOKEN}:x-oauth-basic@github.com/ProjectAthenaa".insteadOf "https://github.com/ProjectAthenaa"
RUN mkdir /app
ADD . /app
WORKDIR /app

#RUN go test -v .

RUN go build -ldflags "-s -w" -o ticket


# final stage
FROM debian:buster-slim
WORKDIR /app
COPY --from=build-env /app/ticket /app/

ENV REDIS_URL=rediss://default:kulqkv6en3um9u09@athena-redis-do-user-9223163-0.b.db.ondigitalocean.com:25061

EXPOSE 3000 3000

ENTRYPOINT ./ticket