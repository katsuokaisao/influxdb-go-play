package influx

import (
	"context"

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
		meas:   "airSensors",
	}
}

func (e *airSensorReader) CheckThreshold10MinutesAgo(ctx context.Context) (*api.QueryTableResult, error) {
	return e.checkThreshold(ctx, "-10m")
}

func (e *airSensorReader) checkThreshold(ctx context.Context, duration string) (*api.QueryTableResult, error) {
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
		Start:          duration,
		TemperatureMax: defaultThreshold.TemperatureMax,
		TemperatureMin: defaultThreshold.TemperatureMin,
		HumidityMax:    defaultThreshold.HumidityMax,
		HumidityMin:    defaultThreshold.HumidityMin,
		Co2Max:         defaultThreshold.Co2Max,
	}
	query := `
		from(bucket: params.bucket)
			|> range(start: params.start)
			|> filter(fn: (r) => r._measurement == params.meas)
			|> filter(fn: (r) => r._field == "temp" or r._field == "hum" or r._field == "co2")
			|> filter(fn: (r) => (r._field == "temp" and (r._value < params.temperature_min or r._value > params.temperature_max)) or
								(r._field == "hum" and (r._value < params.humidity_min or r._value > params.humidity_max)) or
								(r._field == "co2" and r._value > params.co2_max))
			|> yield(name: "exceeded_thresholds")
	`

	return e.cli.QueryWithParams(ctx, query, params)
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

type airSensorDeleter struct {
	cli    api.DeleteAPI
	bucket string
	meas   string
}

func NewAirSensorDeleter(
	cli api.DeleteAPI,
	bucket string,
	meas string,
) domain.AirSensorDeleter {
	return &airSensorDeleter{
		cli:    cli,
		bucket: bucket,
		meas:   "airSensors",
	}
}