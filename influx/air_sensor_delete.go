package influx

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/katsuokaisao/influxdb-play/domain"
)

type airSensorDeleter struct {
	cli    api.DeleteAPI
	org    string
	bucket string
	meas   string
}

func NewAirSensorDeleter(
	cli api.DeleteAPI,
	org string,
	bucket string,
	meas string,
) domain.AirSensorDeleter {
	return &airSensorDeleter{
		cli:    cli,
		org:    org,
		bucket: bucket,
		meas:   meas,
	}
}

func (e *airSensorDeleter) DeleteRecords(ctx context.Context, start time.Time, end time.Time) error {
	predicate := fmt.Sprintf(`_measurement="%s"`, e.meas)
	return e.cli.DeleteWithName(ctx, e.org, e.bucket, start, end, predicate)
}
