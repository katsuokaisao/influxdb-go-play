package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/katsuokaisao/influxdb-play/influx"
	"github.com/spf13/cobra"
)

var get3HourAgoDataPointsCmd = &cobra.Command{
	Use:   "get3HourAgoDataPoints",
	Short: "Get the data points 3 hours ago",
	Run: func(cmd *cobra.Command, args []string) {
		bucket := "get-started"
		org := "example-org"
		token := "yo3yR7t4xC3V1m42EV5djwiXFvYCoSRBF9sDV77QOrezVOxZM9MlqOJkN4uajGcBnrubJfhiis0vijJK7NLFjA=="
		url := "http://localhost:8086"

		client := influxdb2.NewClient(url, token)

		q := client.QueryAPI(org)
		asr := influx.NewAirSensorReader(q, bucket, "home")

		points, err := asr.Get3HoursAgoDataPoints(context.Background())
		if err != nil {
			panic(err)
		}

		b, err := json.MarshalIndent(points, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	},
}
