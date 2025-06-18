# logporter

A simple alternative to [cAdvisor](https://github.com/google/cadvisor) for getting basic Docker container metrics as well as custom metrics (e.g. container uptime and number of logged messages).

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
- [X] Building a Docker image
- [ ] Testing

### Install

<!-- - Download the image from Docker Hub or build it yourself (optional): -->
- Build Docker image:

```bash
git clone https://github.com/Lifailon/logporter
cd logporter
docker build -t logporter:latest .
```

- Run the exporter in the container:

```bash
docker run -d --name logporter \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -p 9333:9333 \
  --restart=unless-stopped \
  logporter:latest
```

- Connect the new target in the `prometheus.yml` configuration:

```yml
scrape_configs:
  - job_name: logporter
    scrape_interval: 5s
    scrape_timeout: 5s
    static_configs:
      - targets:
        - localhost:9333
```

- Import the prepared [Dashboard](cfg/grafana-dashboard.json) into Grafana.

> [!IMPORTANT]
> If you notice an bug in `PromQL` queries in the Grafana Dashboard, please open an new [Issue](https://github.com/Lifailon/logporter/issues) or make the change yourself using a [Pull Request](https://github.com/Lifailon/logporter/pulls).

![](/img/basic-metrics.jpg)

![](/img/other-metrics.jpg)
