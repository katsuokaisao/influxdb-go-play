package influx

import (
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/katsuokaisao/influxdb-play/domain"
)

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
