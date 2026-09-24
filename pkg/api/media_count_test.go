package api

import (
	"encoding/json"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

func TestCountMediaRequestJSON(t *testing.T) {
	body := `{"filter":{"devices":["camera-one"],"timeRanges":[{"start":100,"end":200}]},"limit":1000}`
	var request CountMediaRequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatal(err)
	}
	if request.Limit != 1000 || len(request.Filter.Devices) != 1 || *request.Filter.Devices[0] != "camera-one" {
		t.Fatalf("unexpected request: %+v", request)
	}
	if len(request.Filter.TimeRanges) != 1 || request.Filter.TimeRanges[0].Start != 100 || request.Filter.TimeRanges[0].End != 200 {
		t.Fatalf("unexpected time ranges: %+v", request.Filter.TimeRanges)
	}
}

func TestCountMediaResponseJSON(t *testing.T) {
	for _, test := range []struct {
		name  string
		count models.MediaCount
		want  string
	}{
		{name: "zero", want: `{"count":0,"limitReached":false}`},
		{name: "exact", count: models.MediaCount{Count: 999}, want: `{"count":999,"limitReached":false}`},
		{name: "bounded", count: models.MediaCount{Count: 1000, LimitReached: true}, want: `{"count":1000,"limitReached":true}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := CountMediaSuccessResponse{Data: CountMediaResponse{MediaCount: test.count}}
			body, err := json.Marshal(response)
			if err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(body, &envelope); err != nil {
				t.Fatal(err)
			}
			if string(envelope.Data) != test.want {
				t.Fatalf("data = %s, want %s", envelope.Data, test.want)
			}
			var decoded CountMediaSuccessResponse
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.Data.MediaCount != test.count {
				t.Fatalf("decoded count = %+v, want %+v", decoded.Data.MediaCount, test.count)
			}
		})
	}
}
