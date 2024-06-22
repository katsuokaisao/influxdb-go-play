package influx

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/katsuokaisao/influxdb-play/domain"
)

type airSensorWriterBlocking struct {
	cli  api.WriteAPIBlocking
	meas string
}

func NewAirSensorWriterBlocking(
	cli api.WriteAPIBlocking,
	meas string,
) domain.AirSensorWriterBlocking {
	return &airSensorWriterBlocking{
		cli:  cli,
		meas: meas,
	}
}

func (e *airSensorWriterBlocking) toPoint(a *domain.AirSensor) *write.Point {
	p := influxdb2.NewPointWithMeasurement(e.meas)
	p.AddTag("room", a.Room)
	p.AddField("temperature", a.Temperature)
	p.AddField("humidity", a.Humidity)
	p.AddField("co2", a.CarbonDioxide)
	p.SetTime(a.TS)
	return p
}

func (e *airSensorWriterBlocking) WriteRecord(ctx context.Context, line string) error {
	return e.cli.WriteRecord(ctx, line)
}

func (e *airSensorWriterBlocking) WritePoint(ctx context.Context, a *domain.AirSensor) error {
	p := e.toPoint(a)
	return e.cli.WritePoint(ctx, p)
}

func (e *airSensorWriterBlocking) EnableBatching() {
	e.cli.EnableBatching()
}

func (e *airSensorWriterBlocking) Flush(ctx context.Context) error {
	return e.cli.Flush(ctx)
}
