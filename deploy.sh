#!/bin/bash
set -e

# Delete existing cluster if it exists
kind delete cluster --name jaeger-go-test 2>/dev/null || true

# Create new cluster
kind create cluster --name jaeger-go-test

# Build and load Docker image
docker build -t jaeger-go-test:latest .
kind load docker-image jaeger-go-test:latest --name jaeger-go-test

# Install Contour ingress controller
kubectl apply -f https://projectcontour.io/quickstart/contour.yaml

# Deploy all resources
kubectl apply -f k8s-all.yaml

# Wait for Contour to be ready first
kubectl wait --for=condition=available --timeout=120s deployment/contour -n projectcontour

# Restart Contour to pick up new config
kubectl rollout restart deployment/contour -n projectcontour
kubectl rollout restart daemonset/envoy -n projectcontour

# Wait for deployment to be ready
kubectl wait --for=condition=available --timeout=120s deployment/jaeger-go-test

# Wait for Contour to be ready after restart
kubectl wait --for=condition=available --timeout=120s deployment/contour -n projectcontour

# Wait for Envoy DaemonSet to be ready
kubectl rollout status daemonset/envoy -n projectcontour --timeout=120s

# Start port-forward in background
echo "Starting port-forward on localhost:8080 via Contour..."
kubectl port-forward service/envoy -n projectcontour 8080:80 &
PORT_FORWARD_PID=$!

# Cleanup function to kill processes
cleanup() {
    echo "Stopping port-forward and log streaming..."
    kill $PORT_FORWARD_PID 2>/dev/null || true
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM EXIT

# Stream logs in foreground (this blocks)
echo "Starting Contour log streaming..."
sleep 2

# Stream both Contour and Envoy logs together
(kubectl logs -f deployment/contour -n projectcontour -c contour --prefix=true --tail=20 2>&1 | sed 's/^/[CONTOUR] /' &
 kubectl logs -f daemonset/envoy -n projectcontour -c envoy --prefix=true --tail=20 2>&1 | sed 's/^/[ENVOY] /' &
 wait)