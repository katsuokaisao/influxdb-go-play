package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/katsuokaisao/influxdb-play/domain"
	"github.com/katsuokaisao/influxdb-play/influx"
	"github.com/spf13/cobra"
)

var writeBlockCmd = &cobra.Command{
	Use:   "write_block",
	Short: "Write a new data to InfluxDB using the WriteAPI with the WriteBlock method",
	Run: func(cmd *cobra.Command, args []string) {
		bucket := "get-started"
		org := "example-org"
		token := "yo3yR7t4xC3V1m42EV5djwiXFvYCoSRBF9sDV77QOrezVOxZM9MlqOJkN4uajGcBnrubJfhiis0vijJK7NLFjA=="
		url := "http://localhost:8086"

		client := influxdb2.NewClient(url, token)

		w := client.WriteAPIBlocking(org, bucket)
		asw := influx.NewAirSensorWriterBlocking(w, "home")

		file, err := os.Open("data.txt")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		dataList := make([]domain.AirSensor, 0, 1024)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			sensorData, err := domain.ParseInfluxDBLineToAirSensor(line)
			if err != nil {
				fmt.Println("Error parsing data:", err)
				continue
			}

			dataList = append(dataList, sensorData)
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
		}

		for _, data := range dataList {
			if err := asw.WritePoint(context.TODO(), &data); err != nil {
				fmt.Println("Error writing data:", err)
			}
			fmt.Printf("Wrote data: %v\n", data)
			time.Sleep(1 * time.Second)
		}
	},
}
