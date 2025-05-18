FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
RUN make build

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /app/ctf-ui-operator .

ENTRYPOINT ["/ctf-ui-operator"]
