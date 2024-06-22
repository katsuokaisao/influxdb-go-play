package cmd

import (
	"context"
	"fmt"
	"sort"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/katsuokaisao/influxdb-play/domain"
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

		temperatureOvers := make([]domain.TemperatureOver, 0, len(to))
		for _, v := range to {
			temperatureOvers = append(temperatureOvers, v)
		}
		sort.Slice(temperatureOvers, func(i, j int) bool {
			if temperatureOvers[i].Room == temperatureOvers[j].Room {
				return temperatureOvers[i].TS.Before(temperatureOvers[j].TS)
			}
			return temperatureOvers[i].Room < temperatureOvers[j].Room
		})

		for _, v := range temperatureOvers {
			fmt.Printf("Temperature over: %v\n", v)
		}

		humidityOvers := make([]domain.HumidityOver, 0, len(ho))
		for _, v := range ho {
			humidityOvers = append(humidityOvers, v)
		}
		sort.Slice(humidityOvers, func(i, j int) bool {
			if humidityOvers[i].Room == humidityOvers[j].Room {
				return humidityOvers[i].TS.Before(humidityOvers[j].TS)
			}
			return humidityOvers[i].Room < humidityOvers[j].Room
		})

		for _, v := range humidityOvers {
			fmt.Printf("Humidity over: %v\n", v)
		}

		co2Overs := make([]domain.CarbonDioxideOver, 0, len(co))
		for _, v := range co {
			co2Overs = append(co2Overs, v)
		}
		sort.Slice(co2Overs, func(i, j int) bool {
			if co2Overs[i].Room == co2Overs[j].Room {
				return co2Overs[i].TS.Before(co2Overs[j].TS)
			}
			return co2Overs[i].Room < co2Overs[j].Room
		})

		for _, v := range co2Overs {
			fmt.Printf("Carbon dioxide over: %v\n", v)
		}
	},
}