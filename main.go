package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InfluxURL    string `yaml:"influx_url"`
	InfluxOrg    string `yaml:"influx_org"`
	InfluxBucket string `yaml:"influx_bucket"`
}

func main() {
	config := new(Config).getConf("config.yaml")
	tags := generateTags()
	fields := getMainStats()

	point := write.NewPoint("monitoring-dev", tags, fields, time.Now())

	writeToInflux(config, point)
}

func getMainStats() map[string]interface{} {
	mainStats := make(map[string]interface{})

	cpuCoreCount, err := cpu.Counts(true)
	if err != nil {
		log.Fatal("Error getting CPU count:", err)
	}
	mainStats["cpu_core_count"] = cpuCoreCount

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatal("Error getting CPU usage:", err)
	}
	mainStats["cpu_usage_percent"] = cpuPercent

	cpuPercentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		log.Fatal("Error getting individual CPU percentages:", err)
	}
	for id, value := range cpuPercentages {
		label := fmt.Sprintf("cpu_%s_usage_percent", id)
		mainStats[label] = value
	}

	vmem, err := mem.VirtualMemory()
	if err != nil {
		log.Fatal("Error getting virtual memory", err)
	}
	mainStats["virtual_memory_total"] = vmem.Total
	mainStats["virtual_memory_used"] = vmem.Used
	mainStats["virtual_memory_free"] = vmem.Free

	return mainStats
}

func generateTags() map[string]string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}

	tags := map[string]string{
		"hostname":  hostname,
		"host_os":   runtime.GOOS,
		"host_arch": runtime.GOARCH,
	}
	return tags
}

func (config *Config) getConf(filePath string) *Config {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}
	err = yaml.Unmarshal(yamlFile, config)

	return config
}

func writeToInflux(config *Config, point *write.Point) {
	client := influxdb2.NewClient(config.InfluxURL, os.Getenv("INFLUXDB_TOKEN"))

	org := config.InfluxOrg
	bucket := config.InfluxBucket

	writeAPI := client.WriteAPIBlocking(org, bucket)

	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
		log.Fatal(err)
	}
}
