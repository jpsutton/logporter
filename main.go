package main

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/goforj/godump"
)

type Metrics struct {
	id          []string
	info        []Info
	baseMetrics []BaseMetrics
	logMetrics  []LogMetrics
}

type Info struct {
	id     string
	Name   string
	State  string
	Status string
}

type BaseMetrics struct {
	id                 string
	cpuTotal           float64
	cpuUser            float64
	cpuKernel          float64
	memUsage           int
	memTotal           int
	netReceiveBytes    int
	netReceivePackets  int
	netTransmitBytes   int
	netTransmitPackets int
	ioReadBytes        int
	ioWriteBytes       int
	pids               int
}

type LogMetrics struct {
	id     string
	stdout int
	stderr int
	stdall int
}

// Get information about all containers (second param to get all or only started containers)
func (m *Metrics) getContainers(dockerClient *client.Client, All bool) []Info {
	containers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{All: All})
	if err != nil {
		panic(err)
	}
	var info []Info = []Info{}
	for _, container := range containers {
		// godump.Dump(container)
		var i Info = Info{}
		i.id = container.ID
		i.Name = strings.Replace(container.Names[0], "/", "", 1)
		i.State = container.State
		i.Status = container.Status
		info = append(info, i)
	}
	return info
}

// Get metric list for specified container by id
func (m *Metrics) getBaseMetrics(dockerClient *client.Client, id string) BaseMetrics {
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

	// godump.Dump(data)

	// Extract data and fill structure
	var bm BaseMetrics = BaseMetrics{}
	bm.id = id

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
	memory_usage := memory_stats["usage"].(float64)
	memUsage := int(memory_usage) / 1024 / 1024
	bm.memUsage = memUsage
	memory_limit := memory_stats["limit"].(float64)
	memLimit := int(memory_limit) / 1024 / 1024
	bm.memTotal = memLimit

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

	// Disk
	blkioStats := data["blkio_stats"].(map[string]interface{})
	ioBytesRecursive := blkioStats["io_service_bytes_recursive"].([]interface{})
	for i := range ioBytesRecursive {
		if ioBytesRecursive[i].(map[string]interface{})["op"] == "read" {
			bm.ioReadBytes = int(ioBytesRecursive[i].(map[string]interface{})["value"].(float64))
		} else {
			bm.ioWriteBytes = int(ioBytesRecursive[i].(map[string]interface{})["value"].(float64))
		}
	}

	// Current block IOps
	// int(ioBytesRecursive[i].(map[string]interface{})["value"].(float64)) - bm.ioRead

	// PIDs count
	pidsStats := data["pids_stats"].(map[string]interface{})
	bm.pids = int(pidsStats["current"].(float64))

	return bm
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

	// fmt.Println(string(dataLogs))

	// Convert bytes to text and get array from rows
	lines := strings.Split(string(dataLogs), "\n")

	return len(lines) - 1
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

	// Get a list of containers with status information
	metrics.info = metrics.getContainers(dockerClient, false)

	// Get all container ID array
	for _, i := range metrics.info {
		metrics.id = append(metrics.id, i.id)
	}

	// Get list of basic metrics
	for _, id := range metrics.id {
		metrics.baseMetrics = append(metrics.baseMetrics, metrics.getBaseMetrics(dockerClient, id))
	}

	// Get a list of custom metrics from logs
	for _, id := range metrics.id {
		var stdout int = metrics.getLogsCount(dockerClient, id, true, false)
		var stderr int = metrics.getLogsCount(dockerClient, id, false, true)
		var stdall int = stdout + stderr
		var lm LogMetrics = LogMetrics{
			id:     id,
			stdout: stdout,
			stderr: stderr,
			stdall: stdall,
		}
		metrics.logMetrics = append(metrics.logMetrics, lm)
	}

	godump.Dump(metrics)
}
