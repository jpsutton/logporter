package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	id        string
	cpuTotal  string
	cpuUser   string
	cpuKernel string
}

type LogMetrics struct {
	id     string
	stdout int
	stderr int
	stdall int
}

// Получить информацию о всех контейнерах
// Второй параметр для получения всех или только запущенных контейнеров
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

// Получить список метрик для указанного контейнера по id
func (m *Metrics) getBaseMetrics(dockerClient *client.Client, id string) BaseMetrics {
	stats, err := dockerClient.ContainerStats(context.Background(), id, false)
	if err != nil {
		panic(err)
	}
	defer stats.Body.Close()

	// Читаем статистику
	jsonStats, err := io.ReadAll(stats.Body)
	if err != nil {
		panic(err)
	}

	// Создаем карту из (ключей string и значений с любым типом данных) для извлечения данных из json
	var data map[string]interface{}

	// Парсим json и заполняем карту
	err = json.Unmarshal(jsonStats, &data)
	if err != nil {
		panic(err)
	}

	godump.Dump(data)

	// Извлекаем данные и заполняем структуру
	var bm BaseMetrics = BaseMetrics{}
	bm.id = id

	cpuStats := data["cpu_stats"].(map[string]interface{})
	cpuUsage := cpuStats["cpu_usage"].(map[string]interface{})
	cpuTotal := cpuUsage["total_usage"].(float64)
	// Переводим наносекунд в секунды (деление на 1 000 000 000)
	bm.cpuTotal = fmt.Sprintf("%.3f", cpuTotal/1e9)
	cpuUser := cpuUsage["usage_in_usermode"].(float64)
	bm.cpuUser = fmt.Sprintf("%.3f", cpuUser/1e9)
	cpuKernel := cpuUsage["usage_in_kernelmode"].(float64)
	bm.cpuKernel = fmt.Sprintf("%.3f", cpuKernel/1e9)

	return bm
}

// Получить количество логов для указанного контейнера по id
func (m *Metrics) getLogsCount(dockerClient *client.Client, id string, stdout bool, stderr bool) int {

	// Заполняем параметры для чтения логов контейнеров
	logsOptions := container.LogsOptions{
		ShowStdout: stdout,
		ShowStderr: stderr,
	}

	// Получаем содержимое логов
	logs, err := dockerClient.ContainerLogs(context.Background(), id, logsOptions)
	if err != nil {
		panic(err)
	}
	defer logs.Close()

	// Читаем и парсим json
	dataLogs, err := io.ReadAll(logs)
	if err != nil {
		panic(err)
	}

	// Преобразуем байты в текст и разбиваем его на массив из строк
	lines := strings.Split(string(dataLogs), "\n")

	return len(lines)
}

func main() {
	// Инициализируем основную структуру
	var metrics *Metrics = &Metrics{}

	// Создаем клиент с параметрами подключения из переменных окружения и согласования версии API с Docker daemon
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}
	// Закрывает соединение при выходе из main()
	defer dockerClient.Close()

	// Получаем список контейнеров с информацией о состояние
	metrics.info = metrics.getContainers(dockerClient, false)

	// Извлекаем массив идентификаторов всех контейнеров
	for _, i := range metrics.info {
		metrics.id = append(metrics.id, i.id)
	}

	// Получаем список базовых метрик
	for _, id := range metrics.id {
		metrics.baseMetrics = append(metrics.baseMetrics, metrics.getBaseMetrics(dockerClient, id))
	}

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
