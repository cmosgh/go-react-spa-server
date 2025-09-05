# Optimized Docker Deployment for Go SPA Server

This document outlines the optimized Docker deployment strategy for the Go SPA server, focusing on generating a minimal Docker image suitable for container orchestration platforms like Kubernetes. This image will *only* contain the Go backend. The Single Page Application (SPA) will be served from an external volume mounted at runtime, configured via an environment variable.

## 1. Building the Docker Image

The `Dockerfile` in the project root uses a multi-stage build to create a small, efficient Docker image containing only the Go backend. The SPA static files are expected to be provided via an external volume at runtime.

To build the Docker image, navigate to the project root directory and run the following command:

```bash
docker build -t go-spa-server:latest .
```

Replace `go-spa-server:latest` with your desired image name and tag.

## 2. Pushing to a Docker Registry

After building the image, you can push it to a Docker registry (e.g., Docker Hub, Google Container Registry, AWS ECR) to make it accessible for deployment.

First, tag your image with the registry's address:

```bash
docker tag go-spa-server:latest your-registry/your-username/go-spa-server:latest
```

Then, push the image:

```bash
docker push your-registry/your-username/go-spa-server:latest
```

Ensure you are logged in to your Docker registry (`docker login`).

## 3. Kubernetes Deployment Example

Here's a basic example of how to deploy the application to Kubernetes using a Deployment and a Service. This example demonstrates how to mount an external volume containing your SPA static files and configure the Go server to serve them.

Save these as `deployment.yaml` and `service.yaml`.

### `deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-spa-server
  labels:
    app: go-spa-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: go-spa-server
  template:
    metadata:
      labels:
        app: go-spa-server
    spec:
      containers:
      - name: go-spa-server
        image: your-registry/your-username/go-spa-server:latest # Replace with your image
        ports:
        - containerPort: 8081
        env:
        - name: STATIC_DIR
          value: /app/spa-static # The path inside the container where SPA files will be mounted
        volumeMounts:
        - name: spa-volume
          mountPath: /app/spa-static # Mount point inside the container
      volumes:
      - name: spa-volume
        # This is an example of a hostPath volume. For production, consider
        # PersistentVolumeClaims (PVCs) with appropriate storage classes.
        hostPath:
          path: /path/to/your/spa/dist # Replace with the actual path to your SPA's build output on the host
          type: DirectoryOrCreate
        # Optional: Resource limits and probes for production
        # resources:
        #   limits:
        #     cpu: "500m"
        #     memory: "256Mi"
        #   requests:
        #     cpu: "200m"
        #     memory: "128Mi"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 3
```

### `service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: go-spa-server-service
spec:
  selector:
    app: go-spa-server
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8081
  type: LoadBalancer # Use NodePort or ClusterIP for internal services
```

To deploy to Kubernetes:

```bash
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## 4. Health Checks

The Go SPA server includes a `/healthz` endpoint for health checks. This endpoint returns a `200 OK` status when the server is running and responsive. It can be used by container orchestration platforms like Kubernetes for `livenessProbe` and `readinessProbe` configurations to ensure the application is healthy and ready to receive traffic.

## 5. Verify Image Size

After building the image, you can check its size using:

```bash
docker images go-spa-server:latest
```

**Expected Optimization:** The multi-stage build significantly reduces the final image size by only including the necessary runtime components (the Go backend). This image will be much smaller than one that includes the React frontend.

## 6. Other Container Platforms

The Docker image can also be deployed to other container platforms such as Docker Swarm, Amazon ECS, Google Cloud Run, or Azure Container Instances. The deployment process will vary depending on the platform but will generally involve building the image, pushing it to a registry, and then configuring the platform to pull and run the image, ensuring to mount the SPA static files as an external volume.