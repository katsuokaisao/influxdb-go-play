package influx

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

func ProvideInfluxCli(url, token string) influxdb2.Client {
	return influxdb2.NewClient(url, token)
}

func ProvideWriteAPI(client influxdb2.Client, org, bucket string) api.WriteAPI {
	return client.WriteAPI(org, bucket)
}

func ProvideWriteAPIBlocking(client influxdb2.Client, org, bucket string) api.WriteAPIBlocking {
	return client.WriteAPIBlocking(org, bucket)
}

func ProvideQueryAPI(client influxdb2.Client, org string) api.QueryAPI {
	return client.QueryAPI(org)
}

func ProvideDeleteAPI(client influxdb2.Client) api.DeleteAPI {
	return client.DeleteAPI()
}

func ProvideTaskAPI(client influxdb2.Client) api.TasksAPI {
	return client.TasksAPI()
}
