FROM golang:1.16.0-buster

ENV ELASTIC_APM_SERVER="https://4f8d53c3f63f4c138ff4367b7ebc3967.apm.us-east-1.aws.cloud.es.io:443"
ENV ELASTIC_APM_SERVICE_NAME="Ticket Service"
ENV ELASTIC_APM_SECRET_TOKEN="aBl5cy0EpPbRDLEC6U"
ENV ELASTIC_APM_ENVIRONMENT="DEV"
ENV REDIS_URL="rediss://default:o6f56b0i536gpbr3@test-redis-do-user-9223163-0.b.db.ondigitalocean.com:25061"
ENV PG_URL="postgresql://doadmin:g2p9clpmybup8cq4@test-auth-do-user-9223163-0.b.db.ondigitalocean.com:25060/defaultdb"

RUN mkdir /app

ADD ./src /app

WORKDIR /app

RUN go build .

EXPOSE 3000 3000

CMD["app/main"]