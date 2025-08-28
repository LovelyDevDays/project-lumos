# Ultra-minimal multi-stage Alpine build
# Target: < 50MB with pre-compiled dependencies

# Stage 1: Build dependencies
FROM python:3.13-alpine3.22 AS builder

WORKDIR /build

# Install build dependencies temporarily
RUN apk add --no-cache gcc musl-dev libffi-dev

# Copy and install minimal requirements (only gRPC and protobuf)
COPY cmd/sparse-retrieval-service/requirements-minimal.txt .
RUN pip install --user --no-cache-dir -r requirements-minimal.txt

# Stage 2: Runtime image
FROM python:3.13-alpine3.22

WORKDIR /app

# Install only essential runtime library
RUN apk add --no-cache libstdc++ && rm -rf /var/cache/apk/*

# Copy pre-compiled packages from builder
COPY --from=builder /root/.local/lib /usr/local/lib/

# Copy minimal application code
COPY gen/python/retrieval /app/retrieval
COPY cmd/sparse-retrieval-service/main.py /app/
COPY cmd/sparse-retrieval-service/app /app/app

ENV PYTHONUNBUFFERED=1
ENV GRPC_PORT=50051
ENV QDRANT_HOST=qdrant
ENV QDRANT_PORT=6333
ENV COLLECTION_NAME=jira_bm42_full
ENV MAX_WORKERS=10

# Clean up Python cache
RUN find /usr/local -type f -name "*.pyc" -delete && \
    find /usr/local -type d -name "__pycache__" -delete && \
    find /usr/local -type f -name "*.pyi" -delete && \
    find /usr/local -type f -name "*.pyx" -delete

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONPATH=/app \
    GRPC_PORT=50051

USER nobody
EXPOSE 50051

CMD ["python", "main.py"]