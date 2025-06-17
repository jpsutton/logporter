# logporter

A simple alternative to [cAdvisor](https://github.com/google/cadvisor) for getting basic Docker container metrics as well as custom metrics (e.g. number of logged messages).

![](/img/basic-metrics.jpg)

![](/img/other-metrics.jpg)

> [!IMPORTANT]
> If you notice an bug in `PromQL` queries in the Grafana Dashboard, please open an new [Issue](https://github.com/Lifailon/logporter/issues) or make the change yourself using a [Pull Request](https://github.com/Lifailon/logporter/pulls).

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
- [ ] Building a Docker Image
- [ ] Testing

### Install

- Download the image from Docker Hub and run the exporter in the container:

``

- Connect the new endpoint to the Prometheus configuration:

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