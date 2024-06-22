package influx

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/katsuokaisao/influxdb-play/domain"
)

type airSensorWriter struct {
	cli  api.WriteAPI
	meas string
}

func NewAirSensorWriter(
	cli api.WriteAPI,
	meas string,
) domain.AirSensorWriter {
	return &airSensorWriter{
		cli:  cli,
		meas: meas,
	}
}

func (e *airSensorWriter) WriteRecord(line string) {
	e.cli.WriteRecord(line)
}

func (e *airSensorWriter) toPoint(a *domain.AirSensor) *write.Point {
	p := influxdb2.NewPointWithMeasurement(e.meas)
	p.AddTag("room", a.Room)
	p.AddField("temperature", a.Temperature)
	p.AddField("humidity", a.Humidity)
	p.AddField("co2", a.CarbonDioxide)
	p.SetTime(a.TS)
	return p
}

func (e *airSensorWriter) WritePoint(a *domain.AirSensor) {
	p := e.toPoint(a)
	e.cli.WritePoint(p)
}

func (e *airSensorWriter) Flush() {
	e.cli.Flush()
}

func (e *airSensorWriter) Errors() <-chan error {
	return e.cli.Errors()
}

func (e *airSensorWriter) SetWriteFailedCallback(cb api.WriteFailedCallback) {
	e.cli.SetWriteFailedCallback(cb)
}
