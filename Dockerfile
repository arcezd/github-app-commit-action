# Container image that runs your code
FROM golang:1.22-alpine3.18 AS builder

WORKDIR /out
COPY . .

RUN go build -o action .

# Container image that runs your code
FROM alpine:3.18

# Install required packages
RUN apk add --no-cache git

COPY entrypoint.sh /bin/entrypoint.sh
COPY --from=builder /out/action /bin/action

RUN chmod +x /bin/entrypoint.sh /bin/action

# Code file to execute when the docker container starts up (`entrypoint.sh`)
ENTRYPOINT ["/bin/entrypoint.sh"]