1. Go to https://app.lightstep.com/sandcastle-dev/developer-mode and create a satellite token
2. Update `docker-compose.yaml`
3. Run `docker compose up --build`
4. On another window: `curl localhost:8080/test`
5. Open Jaeger (`http://localhost:16686`) and Lightstep (`https://app.lightstep.com/ACCOUNT/developer-mode`) to view traces

![jaeger 1](img/1.png)

![jaeger 2](img/2.png)