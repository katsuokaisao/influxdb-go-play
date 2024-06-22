package cmd

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/katsuokaisao/influxdb-play/influx"
	"github.com/spf13/cobra"
)

var checkThreshold10MinutesAgoCmd = &cobra.Command{
	Use:   "checkThreshold10MinutesAgo",
	Short: "Check the threshold 10 minutes ago",
	Run: func(cmd *cobra.Command, args []string) {
		bucket := "get-started"
		org := "example-org"
		token := "yo3yR7t4xC3V1m42EV5djwiXFvYCoSRBF9sDV77QOrezVOxZM9MlqOJkN4uajGcBnrubJfhiis0vijJK7NLFjA=="
		url := "http://localhost:8086"

		client := influxdb2.NewClient(url, token)

		q := client.QueryAPI(org)
		asr := influx.NewAirSensorReader(q, bucket, "home")

		to, ho, co, err := asr.CheckThreshold10MinutesAgo(context.Background())
		if err != nil {
			panic(err)
		}

		for _, v := range to {
			fmt.Printf("Temperature over: %v\n", v)
		}

		for _, v := range ho {
			fmt.Printf("Humidity over: %v\n", v)
		}

		for _, v := range co {
			fmt.Printf("Carbon dioxide over: %v\n", v)
		}
	},
}
