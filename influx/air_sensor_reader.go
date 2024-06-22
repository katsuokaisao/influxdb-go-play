package influx

import (
	"context"
	"fmt"
	"sort"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/katsuokaisao/influxdb-play/domain"
)

type airSensorReader struct {
	cli    api.QueryAPI
	bucket string
	meas   string
}

func NewAirSensorReader(
	cli api.QueryAPI,
	bucket string,
	meas string,
) domain.AirSensorReader {
	return &airSensorReader{
		cli:    cli,
		bucket: bucket,
		meas:   meas,
	}
}

func (e *airSensorReader) CheckThreshold10MinutesAgo(ctx context.Context) (
	[]domain.TemperatureOver,
	[]domain.HumidityOver,
	[]domain.CarbonDioxideOver,
	error,
) {
	return e.checkThreshold(ctx, "-10m")
}

func (e *airSensorReader) checkThreshold(ctx context.Context, start string) (
	[]domain.TemperatureOver,
	[]domain.HumidityOver,
	[]domain.CarbonDioxideOver,
	error,
) {
	defaultThreshold := domain.DefaultAirSensorThreshold()
	params := struct {
		Bucket         string  `json:"bucket"`
		Meas           string  `json:"meas"`
		Start          string  `json:"start"`
		TemperatureMax float64 `json:"temperature_max"`
		TemperatureMin float64 `json:"temperature_min"`
		HumidityMax    float64 `json:"humidity_max"`
		HumidityMin    float64 `json:"humidity_min"`
		Co2Max         float64 `json:"co2_max"`
	}{
		Bucket:         e.bucket,
		Meas:           e.meas,
		Start:          start,
		TemperatureMax: defaultThreshold.TemperatureMax,
		TemperatureMin: defaultThreshold.TemperatureMin,
		HumidityMax:    defaultThreshold.HumidityMax,
		HumidityMin:    defaultThreshold.HumidityMin,
		Co2Max:         defaultThreshold.Co2Max,
	}
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s)
			|> filter(fn: (r) => r._measurement == "%s")
			|> filter(fn: (r) => r._field == "temp" or r._field == "hum" or r._field == "co2")
			|> filter(fn: (r) => (r._field == "temp" and (r._value < %f or r._value > %f)) or
								(r._field == "hum" and (r._value < %f or r._value > %f)) or
								(r._field == "co2" and r._value > %f))
			|> yield(name: "exceeded_thresholds")
	`, params.Bucket, params.Start, params.Meas, params.TemperatureMin, params.TemperatureMax, params.HumidityMin, params.HumidityMax, params.Co2Max)
	fmt.Printf("query: %s\n", query)

	result, err := e.cli.Query(ctx, query)
	if err != nil {
		return nil, nil, nil, err
	}

	to := make(map[domain.TemperatureMeta]float64)
	ho := make(map[domain.HumidityMeta]float64)
	co := make(map[domain.CarbonDioxideMeta]float64)
	for result.Next() {
		if result.TableChanged() {
			fmt.Printf("table: %s\n", result.TableMetadata().String())
		}
		r := result.Record()

		if r.Field() == "temperature" {
			t := domain.TemperatureMeta{
				Room: r.ValueByKey("room").(string),
				TS:   r.Time(),
			}
			to[t] = r.Value().(float64)
		}
		if r.Field() == "humidity" {
			h := domain.HumidityMeta{
				Room: r.ValueByKey("room").(string),
				TS:   r.Time(),
			}
			ho[h] = r.Value().(float64)
		}
		if r.Field() == "co2" {
			c := domain.CarbonDioxideMeta{
				Room: r.ValueByKey("room").(string),
				TS:   r.Time(),
			}
			co[c] = r.Value().(float64)
		}
	}
	if result.Err() != nil {
		return nil, nil, nil, result.Err()
	}

	temperatureOvers := make([]domain.TemperatureOver, 0, len(to))
	for k, v := range to {
		temperatureOvers = append(temperatureOvers, domain.TemperatureOver{
			Room:        k.Room,
			TS:          k.TS,
			Temperature: v,
		})
	}
	sort.Slice(temperatureOvers, func(i, j int) bool {
		if temperatureOvers[i].Room == temperatureOvers[j].Room {
			return temperatureOvers[i].TS.Before(temperatureOvers[j].TS)
		}
		return temperatureOvers[i].Room < temperatureOvers[j].Room
	})

	humidityOvers := make([]domain.HumidityOver, 0, len(ho))
	for k, v := range ho {
		humidityOvers = append(humidityOvers, domain.HumidityOver{
			Room:     k.Room,
			TS:       k.TS,
			Humidity: v,
		})
	}
	sort.Slice(humidityOvers, func(i, j int) bool {
		if humidityOvers[i].Room == humidityOvers[j].Room {
			return humidityOvers[i].TS.Before(humidityOvers[j].TS)
		}
		return humidityOvers[i].Room < humidityOvers[j].Room
	})

	co2Overs := make([]domain.CarbonDioxideOver, 0, len(co))
	for k, v := range co {
		co2Overs = append(co2Overs, domain.CarbonDioxideOver{
			Room:          k.Room,
			TS:            k.TS,
			CarbonDioxide: v,
		})
	}
	sort.Slice(co2Overs, func(i, j int) bool {
		if co2Overs[i].Room == co2Overs[j].Room {
			return co2Overs[i].TS.Before(co2Overs[j].TS)
		}
		return co2Overs[i].Room < co2Overs[j].Room
	})

	return temperatureOvers, humidityOvers, co2Overs, nil
}

func (e *airSensorReader) Get3HourAgoDataPoints(ctx context.Context) (*api.QueryTableResult, error) {
	return e.getDataPoints(ctx, "-3h")
}

func (e *airSensorReader) getDataPoints(ctx context.Context, duration string) (*api.QueryTableResult, error) {
	params := struct {
		Bucket string `json:"bucket"`
		Meas   string `json:"meas"`
		Start  string `json:"start"`
	}{
		Bucket: e.bucket,
		Meas:   e.meas,
		Start:  duration,
	}
	query := `
		from(bucket: params.bucket)
			|> range(start: params.start)
			|> filter(fn: (r) => r._measurement == params.meas)
	`

	return e.cli.QueryWithParams(ctx, query, params)
}

func (e *airSensorReader) GetDailyAggregates(ctx context.Context) (*api.QueryTableResult, error) {
	params := struct {
		Bucket string `json:"bucket"`
		Meas   string `json:"meas"`
	}{
		Bucket: e.bucket,
		Meas:   e.meas,
	}
	query := `
		from(bucket: params.bucket)
			|> range(start: -1d)
			|> filter(fn: (r) => r._measurement == params.meas)
			|> aggregateWindow(every: 1d, fn: mean, createEmpty: false)
	`

	return e.cli.QueryWithParams(ctx, query, params)
}
