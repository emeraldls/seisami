
FROM golang:1.22.5-alpine3.19 AS builder

WORKDIR /build
COPY go.mod .
COPY go.sum .

COPY . /build

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build"  go build -o /seisami_backend

FROM alpine:3.14 AS runner

WORKDIR /data

COPY --from=builder /seisami_backend /bin/seisami

CMD ["/bin/seisami"]
