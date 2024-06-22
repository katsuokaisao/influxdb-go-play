package cmd

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/katsuokaisao/influxdb-play/influx"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read data from InfluxDB",
	Run: func(cmd *cobra.Command, args []string) {
		bucket := "get-started"
		org := "example-org"
		token := "yo3yR7t4xC3V1m42EV5djwiXFvYCoSRBF9sDV77QOrezVOxZM9MlqOJkN4uajGcBnrubJfhiis0vijJK7NLFjA=="
		url := "http://localhost:8086"

		client := influxdb2.NewClient(url, token)

		q := client.QueryAPI(org)
		asr := influx.NewAirSensorReader(q, bucket, "home")

		result, err := asr.CheckThreshold10MinutesAgo(context.Background())
		if err != nil {
			panic(err)
		}
		for result.Next() {
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			fmt.Printf("value: %v\n", result.Record().Value())
		}
		if result.Err() != nil {
			panic(result.Err())
		}
	},
}
