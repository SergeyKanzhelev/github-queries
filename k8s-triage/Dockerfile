FROM golang:1.22.0 as builder
WORKDIR /app
RUN ls
COPY * ./
RUN GOOS=linux go build -o /k8s-triage

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /k8s-triage /k8s-triage
ENV PORT 8080
USER nonroot:nonroot
CMD ["/k8s-triage"]
