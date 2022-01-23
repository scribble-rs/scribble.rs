package game

import (
	"encoding/json"
	"testing"

	"github.com/mitchellh/mapstructure"
)

var (
	ordered   = []byte(`{"type":"line","data":{"fromX":1023.53,"fromY":591.06,"toX":1017.65,"toY":602.82,"color":{"r":0,"g":0,"b":0},"lineWidth":8}}`)
	unordered = []byte(`{"data":{"fromX":1023.53,"fromY":591.06,"toX":1017.65,"toY":602.82,"color":{"r":0,"g":0,"b":0},"lineWidth":8},"type":"line"}`)
	mapData   = map[string]interface{}{
		"fromX": 1025.53,
		"fromY": 591.06,
		"toX":   1017.65,
		"toY":   602.82,
		"color": struct {
			r, g, b uint8
		}{},
		"lineWidth": 8,
	}
)

type noData struct {
	Type string `json:"type"`
}

type data struct {
	Type string `json:"type"`
	Data interface{}
}

func BenchmarkUnmarshal_NoData_Ordered(b *testing.B) {
	target := &noData{}
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = json.Unmarshal(ordered, target)
	}
	if err != nil {
		b.Fatal(err.Error())
	}
}

func BenchmarkUnmarshal_Data_Ordered(b *testing.B) {
	target := &data{}
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = json.Unmarshal(ordered, target)
	}
	if err != nil {
		b.Fatal(err.Error())
	}
}

func BenchmarkUnmarshal_NoData_Unordered(b *testing.B) {
	target := &noData{}
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = json.Unmarshal(unordered, target)
	}
	if err != nil {
		b.Fatal(err.Error())
	}
}

func BenchmarkUnmarshal_Data_Unordered(b *testing.B) {
	target := &data{}
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = json.Unmarshal(unordered, target)
	}
	if err != nil {
		b.Fatal(err.Error())
	}
}

func BenchmarkUnmarshal_MapStructure(b *testing.B) {
	target := &LineEvent{}
	var err error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = mapstructure.Decode(mapData, target)
	}
	if err != nil {
		b.Fatal(err.Error())
	}
}

var lineEventSink *LineEvent

func BenchmarkUnmarshal_Reunmarshal(b *testing.B) {
	line := &Line{}
	var err error
	var tempLineEvent *LineEvent
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err = mapstructure.Decode(mapData, line)
		tempLineEvent = &LineEvent{Type: "type", Data: line}
	}
	lineEventSink = tempLineEvent
	if err != nil {
		b.Fatal(err.Error())
	}
}
