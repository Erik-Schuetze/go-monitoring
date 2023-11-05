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
	fields := make(map[string]interface{})

	fields["test_value"] = 1
	point := write.NewPoint("monitoring-dev", tags, fields, time.Now())

	writeToInflux(config, point)
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
