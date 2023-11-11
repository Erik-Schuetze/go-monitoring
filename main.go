package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
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

	hostInfo, err := host.Info()
	if err != nil {
		log.Fatal("Error getting host info:", err)
	}
	mainStats["uptime"] = hostInfo.Uptime

	loadInfo, err := load.Avg()
	if err != nil {
		log.Fatal("Error getting CPU load average:", err)
	}
	mainStats["load_avg_1"] = loadInfo.Load1
	mainStats["load_avg_5"] = loadInfo.Load5
	mainStats["load_avg_15"] = loadInfo.Load15

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
		label := fmt.Sprintf("cpu_%d_usage_percent", id)
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
	info, err := host.Info()
	if err != nil {
		fmt.Println(err)
	}

	tags := map[string]string{
		"hostname":             info.Hostname,
		"host_os":              info.OS,
		"host_platform":        info.Platform,
		"host_platform_family": info.PlatformFamily,
		"host_id":              info.HostID,
		"host_arch":            info.KernelArch,
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
