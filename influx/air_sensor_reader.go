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
			|> sort(columns: ["room", "_time"], desc: false)
			|> yield(name: "exceeded_thresholds")
	`, params.Bucket, params.Start, params.Meas, params.TemperatureMin, params.TemperatureMax, params.HumidityMin, params.HumidityMax, params.Co2Max)
	fmt.Printf("query: %s\n", query)

	result, err := e.cli.Query(ctx, query)
	if err != nil {
		return nil, nil, nil, err
	}

	temperatureOvers := make([]domain.TemperatureOver, 0, 1024)
	humidityOvers := make([]domain.HumidityOver, 0, 1024)
	co2Overs := make([]domain.CarbonDioxideOver, 0, 1024)
	for result.Next() {
		r := result.Record()
		if r.Field() == "temperature" {
			t := domain.TemperatureOver{
				Room:        r.ValueByKey("room").(string),
				Temperature: r.Value().(float64),
				TS:          r.Time(),
			}
			temperatureOvers = append(temperatureOvers, t)
		}
		if r.Field() == "humidity" {
			h := domain.HumidityOver{
				Room:     r.ValueByKey("room").(string),
				Humidity: r.Value().(float64),
				TS:       r.Time(),
			}
			humidityOvers = append(humidityOvers, h)
		}
		if r.Field() == "co2" {
			c := domain.CarbonDioxideOver{
				Room:          r.ValueByKey("room").(string),
				CarbonDioxide: r.Value().(float64),
				TS:            r.Time(),
			}
			co2Overs = append(co2Overs, c)
		}
	}
	if result.Err() != nil {
		return nil, nil, nil, result.Err()
	}

	return temperatureOvers, humidityOvers, co2Overs, nil
}

func (e *airSensorReader) Get3HoursAgoDataPoints(ctx context.Context) ([]domain.AirSensor, error) {
	return e.getDataPoints(ctx, "-3h")
}

func (e *airSensorReader) getDataPoints(ctx context.Context, duration string) ([]domain.AirSensor, error) {
	params := struct {
		Bucket string `json:"bucket"`
		Meas   string `json:"meas"`
		Start  string `json:"start"`
	}{
		Bucket: e.bucket,
		Meas:   e.meas,
		Start:  duration,
	}

	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: %s)
			|> filter(fn: (r) => r._measurement == "%s")
	`, params.Bucket, params.Start, params.Meas)

	result, err := e.cli.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	airSensors := make(map[domain.AirSensorMeta]domain.AirSensor)
	for result.Next() {
		r := result.Record()
		meta := domain.AirSensorMeta{
			Room: r.ValueByKey("room").(string),
			TS:   r.Time(),
		}
		if _, ok := airSensors[meta]; !ok {
			airSensors[meta] = domain.AirSensor{
				Room: meta.Room,
				TS:   meta.TS,
			}
		}
		switch r.Field() {
		case "temp":
			airSensors[meta] = domain.AirSensor{
				Room:          meta.Room,
				TS:            meta.TS,
				Temperature:   r.Value().(float64),
				Humidity:      airSensors[meta].Humidity,
				CarbonDioxide: airSensors[meta].CarbonDioxide,
			}
		case "hum":
			airSensors[meta] = domain.AirSensor{
				Room:          meta.Room,
				TS:            meta.TS,
				Temperature:   airSensors[meta].Temperature,
				Humidity:      r.Value().(float64),
				CarbonDioxide: airSensors[meta].CarbonDioxide,
			}
		case "co2":
			airSensors[meta] = domain.AirSensor{
				Room:          meta.Room,
				TS:            meta.TS,
				Temperature:   airSensors[meta].Temperature,
				Humidity:      airSensors[meta].Humidity,
				CarbonDioxide: r.Value().(float64),
			}
		}
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	airSensorsSlice := make([]domain.AirSensor, 0, len(airSensors))
	for _, v := range airSensors {
		airSensorsSlice = append(airSensorsSlice, v)
	}
	sort.Slice(airSensorsSlice, func(i, j int) bool {
		if airSensorsSlice[i].Room == airSensorsSlice[j].Room {
			return airSensorsSlice[i].TS.Before(airSensorsSlice[j].TS)
		}
		return airSensorsSlice[i].Room < airSensorsSlice[j].Room
	})

	return airSensorsSlice, nil
}

func (e *airSensorReader) GetDailyAggregates(ctx context.Context) (*api.QueryTableResult, error) {
	params := struct {
		Bucket string `json:"bucket"`
		Meas   string `json:"meas"`
	}{
		Bucket: e.bucket,
		Meas:   e.meas,
	}

	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -1d)
			|> filter(fn: (r) => r._measurement == "%s")
			|> aggregateWindow(every: 1h, fn: mean, createEmpty: false)
	`, params.Bucket, params.Meas)

	return e.cli.Query(ctx, query)
}
