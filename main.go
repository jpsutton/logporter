package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Metrics struct {
	id          []string
	info        map[string]*Info
	baseMetrics map[string]*BaseMetrics
	logMetrics  map[string]*LogMetrics
}

type Info struct {
	name   string
	state  string
	status string
}

type BaseMetrics struct {
	cpuTotal           float64
	cpuUser            float64
	cpuKernel          float64
	memTotalBtyes      int
	memUsageBtyes      int
	netReceiveBytes    int
	netReceivePackets  int
	netTransmitBytes   int
	netTransmitPackets int
	ioReadBytes        int
	ioWriteBytes       int
	pids               int
}

type LogMetrics struct {
	stdout int
	stderr int
	stdall int
}

// Get information about all containers (second param to get all or only started containers)
func (m *Metrics) getContainers(dockerClient *client.Client, All bool) (map[string]*Info, []string) {
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{All: All})
	if err != nil {
		panic(err)
	}
	info := map[string]*Info{}
	var idArr []string
	for _, container := range containers {
		// Debug output container info
		// godump.Dump(container)

		i := Info{}
		currentId := container.ID
		i.name = strings.Replace(container.Names[0], "/", "", 1)
		i.state = container.State
		i.status = container.Status
		info[currentId] = &i
		idArr = append(idArr, currentId)
	}
	return info, idArr
}

// Get metric list for specified container by id
func (m *Metrics) getBaseMetrics(dockerClient *client.Client, id string) *BaseMetrics {
	stats, err := dockerClient.ContainerStats(context.Background(), id, false)
	if err != nil {
		panic(err)
	}
	defer stats.Body.Close()

	// Read statistics
	jsonStats, err := io.ReadAll(stats.Body)
	if err != nil {
		panic(err)
	}

	// Create a map to extract data from json
	var data map[string]interface{}

	// Parsing json and fill in map
	err = json.Unmarshal(jsonStats, &data)
	if err != nil {
		panic(err)
	}

	// Debug output metrics
	// godump.Dump(data)

	// Extract data and fill structure
	var bm BaseMetrics = BaseMetrics{}

	// Processor
	cpuStats := data["cpu_stats"].(map[string]interface{})
	cpuUsage := cpuStats["cpu_usage"].(map[string]interface{})
	cpuTotal := cpuUsage["total_usage"].(float64)
	// Convert nanoseconds to seconds (divided by 1 000 000 000 000)
	bm.cpuTotal = cpuTotal / 1e9
	cpuUser := cpuUsage["usage_in_usermode"].(float64)
	bm.cpuUser = cpuUser / 1e9
	cpuKernel := cpuUsage["usage_in_kernelmode"].(float64)
	bm.cpuKernel = cpuKernel / 1e9

	// Memory
	memory_stats := data["memory_stats"].(map[string]interface{})
	memory_limit := memory_stats["limit"].(float64)
	memLimit := int(memory_limit)
	bm.memTotalBtyes = memLimit
	memory_usage := memory_stats["usage"].(float64)
	memUsage := int(memory_usage)
	bm.memUsageBtyes = memUsage

	// Network
	networks := data["networks"].(map[string]interface{})
	network_interface := networks["eth0"].(map[string]interface{})
	rx_bytes := network_interface["rx_bytes"].(float64)
	bm.netReceiveBytes = int(rx_bytes)
	rx_packets := network_interface["rx_packets"].(float64)
	bm.netReceivePackets = int(rx_packets)
	tx_bytes := network_interface["tx_bytes"].(float64)
	bm.netTransmitBytes = int(tx_bytes)
	tx_packets := network_interface["tx_packets"].(float64)
	bm.netTransmitPackets = int(tx_packets)

	// IO
	blkioStats := data["blkio_stats"].(map[string]interface{})
	ioBytesRecursive := blkioStats["io_service_bytes_recursive"].([]interface{})
	for i := range ioBytesRecursive {
		if ioBytesRecursive[i].(map[string]interface{})["op"] == "read" {
			bm.ioReadBytes = int(ioBytesRecursive[i].(map[string]interface{})["value"].(float64))
		} else {
			bm.ioWriteBytes = int(ioBytesRecursive[i].(map[string]interface{})["value"].(float64))
		}
	}

	// PIDs count
	pidsStats := data["pids_stats"].(map[string]interface{})
	bm.pids = int(pidsStats["current"].(float64))

	return &bm
}

// Get line count from logs for specified container by id
func (m *Metrics) getLogsCount(dockerClient *client.Client, id string, stdout bool, stderr bool) int {

	// Fill in options to read container logs
	logsOptions := container.LogsOptions{
		ShowStdout: stdout,
		ShowStderr: stderr,
	}

	// Get log content
	logs, err := dockerClient.ContainerLogs(context.Background(), id, logsOptions)
	if err != nil {
		panic(err)
	}
	defer logs.Close()

	// Read and parsing json
	dataLogs, err := io.ReadAll(logs)
	if err != nil {
		panic(err)
	}

	// Debug output logs
	// fmt.Println(string(dataLogs))

	// Convert bytes to text and get array from rows
	lines := strings.Split(string(dataLogs), "\n")

	countLogs := len(lines) - 1

	return countLogs
}

// Converting metrics to Prometheus format
func (m *Metrics) prometheusFormat(metricName, helpText, typeData, id, containerName, hostname string, value any) []string {
	var metricsText []string

	metricsText = append(metricsText, "# HELP "+metricName+" "+helpText)
	metricsText = append(metricsText, "# TYPE "+metricName+" "+typeData)
	metricsLine := fmt.Sprintf(
		"%s{containerId=\"%s\",containerName=\"%s\",hostname=\"%s\"} %v",
		metricName, id, containerName, hostname, value,
	)
	metricsText = append(metricsText, metricsLine)

	return metricsText
}

// Getting all metrics in Prometheus format
func (m *Metrics) prometheusMetrics(id string) []string {
	// Main text slice
	var prometheusMetrics []string

	// Get hostname
	hostname, _ := os.Hostname()

	// Get container name
	containerName := m.info[id].name

	// Processor
	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_cpu_usage_total",
		"Total CPU usage (user and kernel) in seconds",
		"gauge", id, containerName, hostname,
		m.baseMetrics[id].cpuTotal,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_cpu_usage_user",
		"User CPU usage in seconds",
		"gauge", id, containerName, hostname,
		m.baseMetrics[id].cpuUser,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_cpu_usage_kernel",
		"Kernel CPU usage in seconds",
		"gauge", id, containerName, hostname,
		m.baseMetrics[id].cpuKernel,
	)...)

	// Memory
	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_memory_total",
		"Total memory size in bytes",
		"gauge", id, containerName, hostname,
		m.baseMetrics[id].memTotalBtyes,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_memory_usage",
		"Usage memory size in bytes",
		"gauge", id, containerName, hostname,
		m.baseMetrics[id].memUsageBtyes,
	)...)

	// Network
	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_network_received_bytes",
		"Number of bytes received on the network",
		"counter", id, containerName, hostname,
		m.baseMetrics[id].netReceiveBytes,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_network_received_packages",
		"Number of packages received on the network",
		"counter", id, containerName, hostname,
		m.baseMetrics[id].netReceivePackets,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_network_transmit_bytes",
		"Number of bytes transmitted on the network",
		"counter", id, containerName, hostname,
		m.baseMetrics[id].netTransmitBytes,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_network_transmit_packages",
		"Number of packages transmitted on the network",
		"counter", id, containerName, hostname,
		m.baseMetrics[id].netTransmitPackets,
	)...)

	// IO
	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_io_read_bytes",
		"Number of bytes read by the block device",
		"counter", id, containerName, hostname,
		m.baseMetrics[id].ioReadBytes,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_io_write_bytes",
		"Number of bytes write by the block device",
		"counter", id, containerName, hostname,
		m.baseMetrics[id].ioWriteBytes,
	)...)

	// PIDs
	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_process_pids_count",
		"Number of running processes and threads",
		"gauge", id, containerName, hostname,
		m.baseMetrics[id].pids,
	)...)

	// Logs
	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_logs_stdout_count",
		"Number of logs from stdout stream in the last 10 seconds",
		"counter", id, containerName, hostname,
		m.logMetrics[id].stdout,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_logs_stderr_count",
		"Number of logs from stderr stream in the last 10 seconds",
		"counter", id, containerName, hostname,
		m.logMetrics[id].stderr,
	)...)

	prometheusMetrics = append(prometheusMetrics, m.prometheusFormat(
		"docker_logs_all_count",
		"Number of logs from all stream in the last 10 seconds",
		"counter", id, containerName, hostname,
		m.logMetrics[id].stdall,
	)...)

	prometheusMetrics = append(prometheusMetrics, "")

	return prometheusMetrics
}

func main() {
	// Initialize the main structure
	var metrics *Metrics = &Metrics{}

	// Create client with connection parameters from environment variables and approval of the API version with the Docker Daemon
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}
	defer dockerClient.Close()

	// /metrics endpoint
	// http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
	// Get a list of containers with status information and all container ID array
	metrics.info, metrics.id = metrics.getContainers(dockerClient, false)

	// Get list of basic metrics
	metrics.baseMetrics = map[string]*BaseMetrics{}
	for _, id := range metrics.id {
		metrics.baseMetrics[id] = metrics.getBaseMetrics(dockerClient, id)
	}

	// // Get a list of custom metrics from logs
	metrics.logMetrics = map[string]*LogMetrics{}
	for _, id := range metrics.id {
		stdout := metrics.getLogsCount(dockerClient, id, true, false)
		stderr := metrics.getLogsCount(dockerClient, id, false, true)
		stdall := stdout + stderr
		var lm LogMetrics = LogMetrics{
			stdout: stdout,
			stderr: stderr,
			stdall: stdall,
		}
		metrics.logMetrics[id] = &lm
	}

	// Debug output main structure
	// godump.Dump(metrics)

	// Get metrics in Prometheus format
	var prometheusMetrics []string
	for _, id := range metrics.id {
		prometheusMetrics = append(prometheusMetrics, metrics.prometheusMetrics(id)...)
	}

	// w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Output metrics in Prometheus format
	for _, m := range prometheusMetrics {
		// fmt.Fprintln(w, m)
		// Debug output metrics on console
		fmt.Println(m)
	}
	// })

	// Start HTTP server
	http.ListenAndServe(":8080", nil)
}
