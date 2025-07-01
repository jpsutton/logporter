# logporter

A simple and lightweight alternative to [cAdvisor](https://github.com/google/cadvisor) for getting basic and custom Docker containers metrics (e.g. container uptime and number of logged messages).

## Why collect log counts?

- Compare the increased load with the number of logged messages.
- If the application is not resource-intensive, the number of logged messages will show the increased load.
- Speed ​​up log analysis by displaying at what point in time the most messages were received from the standard and error stream.

## Roadmap

- [x] Functions for extracting basic and custom metrics
- [x] Converting metrics to Prometheus format
- [X] HTTP server and logging
- [X] Error handling (check for missing data)
- [X] Getting data in a goroutine
- [X] Grafana Dashboard
- [X] Docker image
- [ ] Testing

## Build

Build the Docker image yourself (optional):

```bash
git clone https://github.com/Lifailon/logporter
cd logporter
docker build -t lifailon/logporter .
# or build for different architectures
docker buildx build --platform linux/amd64,linux/arm64 .
```

## Install

- Run the exporter in a container using an image from [Docker Hub](https://hub.docker.com/r/lifailon/logporter):

```bash
docker run -d --name logporter \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -p 9333:9333 \
  --restart=unless-stopped \
  lifailon/logporter:latest
```

- Connect the new target in the `prometheus.yml` configuration:

```yml
scrape_configs:
  - job_name: logporter
    scrape_interval: 10s
    scrape_timeout: 10s
    static_configs:
      - targets:
        - localhost:9333
```

- Import the prepared public [Grafana Dashboard](https://grafana.com/grafana/dashboards/23573-docker-exporter-logporter) using the id `23573` or from a [json](https://github.com/Lifailon/logporter/blob/main/cfg/grafana-dashboard.json) file.

> [!IMPORTANT]
> If you notice a bug in `PromQL` queries or want to improve the Grafana dashboard, create a new [issue](https://github.com/Lifailon/logporter/issues) or make the changes yourself using a [Pull Request](https://github.com/Lifailon/logporter/pulls).

![](/img/basic-metrics.jpg)

![](/img/other-metrics.jpg)
