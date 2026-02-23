# container image for building binary
FROM golang:1.26.0-alpine3.23 AS builder

WORKDIR /out
COPY . .

RUN go build -o action .

# container image that runs your code
FROM alpine:3.23

# install required packages
RUN apk add --no-cache git

COPY entrypoint.sh /bin/entrypoint.sh
COPY --from=builder /out/action /bin/action

RUN chmod +x /bin/entrypoint.sh /bin/action

# code file to execute when the docker container starts up (`entrypoint.sh`)
ENTRYPOINT ["/bin/entrypoint.sh"]