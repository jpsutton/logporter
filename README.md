# logporter

Prometheus exporter for getting base metrics and number of logged messages in containers.

## Why log the number of logs?

- Display the number of requests to the application if the application registers each request by keywords.
- Compare the increased load with the number of logged messages.
- If the application is not resource-intensive, the number of logged messages will show the increased load.
- Speed ​​up log analysis by displaying at what point in time the most messages were received from the standard and error stream.

## Roadmap

- [x] Functions for extracting base and custom metrics
- [x] Converting metrics to Prometheus format
- [X] HTTP server
- [X] Error handling (check for missing data)
- [ ] Getting data in a goroutine
- [ ] Building a Docker image
- [ ] Grafana dashboard
- [ ] Testing
