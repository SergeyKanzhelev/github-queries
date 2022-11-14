FROM golang:1.19.2 as builder
WORKDIR /app
RUN ls
COPY * ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /k8s-triage

FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /k8s-triage /k8s-triage
ENV PORT 8080
USER nonroot:nonroot
CMD ["/k8s-triage"]
