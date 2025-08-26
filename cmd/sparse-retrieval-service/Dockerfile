# Stage 1: Proto generation
FROM python:3.12-slim AS proto-builder

WORKDIR /build

# Install protobuf compiler and tools  
RUN apt-get update && apt-get install -y \
    protobuf-compiler \
    && rm -rf /var/lib/apt/lists/* \
    && pip install --no-cache-dir grpcio-tools

# Copy proto files
COPY proto/ ./proto/

# Generate Python protobuf code
RUN mkdir -p gen/python && \
    python -m grpc_tools.protoc \
    -I./proto \
    --python_out=gen/python \
    --grpc_python_out=gen/python \
    --pyi_out=gen/python \
    proto/retrieval/passage/v1/*.proto \
    proto/retrieval/issue/v1/*.proto

# Stage 2: Python dependencies builder
FROM python:3.13-slim AS deps-builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy and install Python dependencies
COPY cmd/sparse-retrieval-service/requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# Stage 3: Final runtime image
FROM python:3.13-slim

WORKDIR /app

# Install only runtime dependencies (minimal)
RUN apt-get update && apt-get install -y \
    libgomp1 \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy Python packages from deps-builder
COPY --from=deps-builder /root/.local /root/.local

# Copy generated protobuf code
COPY --from=proto-builder /build/gen ./gen

# Copy application code
COPY cmd/sparse-retrieval-service/ ./cmd/sparse-retrieval-service/

# Make sure scripts are in PATH
ENV PATH=/root/.local/bin:$PATH
ENV PYTHONPATH=/app:/app/gen/python:/root/.local/lib/python3.13/site-packages

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV GRPC_PORT=50051
ENV QDRANT_HOST=qdrant
ENV QDRANT_PORT=6333
ENV COLLECTION_NAME=jira_bm42_full
ENV MAX_WORKERS=10

# Create non-root user for security (disabled for now)
# RUN useradd -m -u 1000 appuser && \
#     chown -R appuser:appuser /app
# USER appuser

# Expose gRPC port
EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import grpc; channel = grpc.insecure_channel('localhost:50051'); exit(0 if channel.channel_ready() else 1)" || exit 1

# Run the server
CMD ["python", "cmd/sparse-retrieval-service/main.py"]