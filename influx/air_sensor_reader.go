package influx

import (
	"context"
	"fmt"

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
	map[domain.TemperatureMeta]domain.TemperatureOver,
	map[domain.HumidityMeta]domain.HumidityOver,
	map[domain.CarbonDioxideOver]domain.CarbonDioxideOver,
	error,
) {
	return e.checkThreshold(ctx, "-10m")
}

func (e *airSensorReader) checkThreshold(ctx context.Context, start string) (
	map[domain.TemperatureMeta]domain.TemperatureOver,
	map[domain.HumidityMeta]domain.HumidityOver,
	map[domain.CarbonDioxideOver]domain.CarbonDioxideOver,
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

	to := make(map[domain.TemperatureMeta]domain.TemperatureOver)
	ho := make(map[domain.HumidityMeta]domain.HumidityOver)
	co := make(map[domain.CarbonDioxideOver]domain.CarbonDioxideOver)
	for result.Next() {
		if result.TableChanged() {
			fmt.Printf("table: %s\n", result.TableMetadata().String())
		}
		r := result.Record()

		if r.Field() == "temperature" {
			t := domain.TemperatureMeta{
				Room:        r.ValueByKey("room").(string),
				Temperature: r.Value().(float64),
			}
			to[t] = domain.TemperatureOver{
				Room:        r.ValueByKey("room").(string),
				Temperature: r.Value().(float64),
				TS:          r.Time(),
			}
		}
		if r.Field() == "humidity" {
			h := domain.HumidityMeta{
				Room:     r.ValueByKey("room").(string),
				Humidity: r.Value().(float64),
			}
			ho[h] = domain.HumidityOver{
				Room:     r.ValueByKey("room").(string),
				Humidity: r.Value().(float64),
				TS:       r.Time(),
			}
		}
		if r.Field() == "co2" {
			c := domain.CarbonDioxideOver{
				Room:          r.ValueByKey("room").(string),
				CarbonDioxide: r.Value().(float64),
				TS:            r.Time(),
			}
			co[c] = c
		}
	}
	if result.Err() != nil {
		return nil, nil, nil, result.Err()
	}

	return to, ho, co, nil
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
