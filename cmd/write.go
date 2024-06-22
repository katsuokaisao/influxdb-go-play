package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/katsuokaisao/influxdb-play/domain"
	"github.com/katsuokaisao/influxdb-play/influx"
	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write a new data to InfluxDB",
	Run: func(cmd *cobra.Command, args []string) {
		bucket := "get-started"
		org := "example-org"
		token := "yo3yR7t4xC3V1m42EV5djwiXFvYCoSRBF9sDV77QOrezVOxZM9MlqOJkN4uajGcBnrubJfhiis0vijJK7NLFjA=="
		url := "http://localhost:8086"

		client := influxdb2.NewClient(url, token)

		w := client.WriteAPI(org, bucket)
		asw := influx.NewAirSensorWriter(w, "airSensors")

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

		errCh := asw.Errors()
		go func() {
			for err := range errCh {
				fmt.Printf("Error writing data: %v\n", err)
			}
		}()

		for _, data := range dataList {
			data.TS = time.Now()
			asw.WritePoint(&data)
			fmt.Printf("Wrote data: %v\n", data)
			time.Sleep(1 * time.Second)
		}

		client.Close()
	},
}
