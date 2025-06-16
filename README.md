# logporter

A simple alternative to [cAdvisor](https://github.com/google/cadvisor) for getting basic metrics from Docker containers, as well as custom metrics (e.g. number of logged messages).

## Why log the number of logs?

- Compare the increased load with the number of logged messages.
- If the application is not resource-intensive, the number of logged messages will show the increased load.
- Speed ​​up log analysis by displaying at what point in time the most messages were received from the standard and error stream.

## Roadmap

- [x] Functions for extracting base and custom metrics
- [x] Converting metrics to Prometheus format
- [X] HTTP server and logging
- [X] Error handling (check for missing data)
- [X] Getting data in a goroutine
- [ ] Grafana dashboard
- [ ] Building a Docker image
- [ ] Add other custom metrics to track file and configuration changes (`diff` and `inspect` methods)
- [ ] Testing
