# logporter

Prometheus exporter for getting base metrics and number of logged messages in containers.

## For what?

- Comparison of increased load with the number of logged messages.
- If the application is not resource-intensive, the exporter will display increased load by the number of logged messages. Experimentally, you can determine threshold values ​​for your applications under average or increased load.
- Preventing incidents and accelerating analysis. For example, unexpected changes occurred in the application's operation, so when analyzing logs, it will be useful to know at what point in time the most messages and errors were received.

## Roadmap

- [x] Functions for extracting base and custom metrics
- [x] Converting metrics to Prometheus format
- [X] HTTP server
- [ ] Logging the exporter work
- [ ] Error handling
- [ ] Getting data in a goroutine
- [ ] Build Docker image
- [ ] Dashboard Grafana
- [ ] Testing
