package domain

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
)

type AirSensor struct {
	Room          string
	Temperature   float64
	Humidity      float64
	CarbonDioxide float64
	TS            time.Time
}

type AirSensorMeta struct {
	Room string
	TS   time.Time
}

type TemperatureOver struct {
	Room        string
	Temperature float64
	TS          time.Time
}

func (t *TemperatureOver) String() string {
	return fmt.Sprintf("Room: %s, Temperature: %f, TS: %s", t.Room, t.Temperature, t.TS)
}

type HumidityOver struct {
	Room     string
	Humidity float64
	TS       time.Time
}

func (h *HumidityOver) String() string {
	return fmt.Sprintf("Room: %s, Humidity: %f, TS: %s", h.Room, h.Humidity, h.TS)
}

type CarbonDioxideOver struct {
	Room          string
	CarbonDioxide float64
	TS            time.Time
}

func (c *CarbonDioxideOver) String() string {
	return fmt.Sprintf("Room: %s, CarbonDioxide: %f, TS: %s", c.Room, c.CarbonDioxide, c.TS)
}

func ParseInfluxDBLineToAirSensor(line string) (AirSensor, error) {
	parts := splitInfluxDBLine(line)
	if len(parts) != 3 {
		return AirSensor{}, fmt.Errorf("invalid data format")
	}

	firstStr := parts[0]
	fieldStr := parts[1]
	timestampStr := parts[2]

	first := strings.Split(firstStr, ",")
	// measurement := first[0]
	tagsPart := strings.Join(first[1:], ",")

	tags := parseTags(tagsPart)
	fields := parseFields(fieldStr)

	timeInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return AirSensor{}, fmt.Errorf("invalid timestamp: %v", err)
	}
	timestampTime := time.Unix(timeInt, 0)

	sensorData := AirSensor{
		Room:          tags["room"],
		Temperature:   fields["temp"].(float64),
		Humidity:      fields["hum"].(float64),
		CarbonDioxide: fields["co"].(float64),
		TS:            timestampTime,
	}

	return sensorData, nil
}

func splitInfluxDBLine(line string) []string {
	parts := []string{}
	var buffer bytes.Buffer
	escaping := false

	for _, char := range line {
		switch char {
		case '\\':
			if escaping {
				buffer.WriteRune(char)
				escaping = false
			} else {
				escaping = true
			}
		case ' ':
			if escaping {
				buffer.WriteRune(char)
				escaping = false
			} else {
				parts = append(parts, buffer.String())
				buffer.Reset()
			}
		default:
			if escaping {
				buffer.WriteRune('\\')
				escaping = false
			}
			buffer.WriteRune(char)
		}
	}
	parts = append(parts, buffer.String())
	return parts
}

func parseTags(tagsPart string) map[string]string {
	tags := map[string]string{}
	tagsArray := strings.Split(tagsPart, ",")
	for _, tag := range tagsArray {
		parts := strings.Split(tag, "=")
		key := parts[0]
		value := parts[1]
		tags[key] = value
	}
	return tags
}

func parseFields(fieldsPart string) map[string]interface{} {
	fields := map[string]interface{}{}
	fieldsArray := strings.Split(fieldsPart, ",")
	for _, field := range fieldsArray {
		parts := strings.Split(field, "=")
		key := parts[0]
		value := parts[1]
		fields[key] = parseFieldValue(value)
	}
	return fields
}

func parseFieldValue(value string) interface{} {
	// 整数と浮動小数点数を分けてパース
	if value[len(value)-1] == 'i' {
		intValue, _ := strconv.ParseInt(value[:len(value)-1], 10, 64)
		return float64(intValue)
	}
	floatValue, _ := strconv.ParseFloat(value, 64)
	return floatValue
}

type AirSensorThreshold struct {
	TemperatureMax float64 `json:"temperature_max"`
	TemperatureMin float64 `json:"temperature_min"`
	HumidityMax    float64 `json:"humidity_max"`
	HumidityMin    float64 `json:"humidity_min"`
	Co2Max         float64 `json:"co2_max"`
}

func DefaultAirSensorThreshold() *AirSensorThreshold {
	return &AirSensorThreshold{
		TemperatureMax: 28,
		TemperatureMin: 20,
		HumidityMax:    60,
		HumidityMin:    40,
		Co2Max:         1000,
	}
}

type AirSensorWriter interface {
	WriteRecord(line string)
	WritePoint(a *AirSensor)
	Flush()
	Errors() <-chan error
	SetWriteFailedCallback(cb api.WriteFailedCallback)
}

type AirSensorWriterBlocking interface {
	WriteRecord(ctx context.Context, line string) error
	WritePoint(ctx context.Context, a *AirSensor) error
	Flush(ctx context.Context) error
	EnableBatching()
}

type AirSensorReader interface {
	CheckThreshold10MinutesAgo(ctx context.Context) (
		[]TemperatureOver,
		[]HumidityOver,
		[]CarbonDioxideOver,
		error,
	)
	Get3HourAgoDataPoints(ctx context.Context) (*api.QueryTableResult, error)
	GetDailyAggregates(ctx context.Context) (*api.QueryTableResult, error)
}

type AirSensorDeleter interface {
}

type AirSensorTask interface {
}
