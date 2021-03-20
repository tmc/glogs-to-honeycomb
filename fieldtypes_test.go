package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_fixFieldTypes(t *testing.T) {
	type args struct {
		l CloudLoggingPayload
	}
	tests := []struct {
		name string
		args args
		want CloudLoggingPayload
	}{
		{"empty", args{CloudLoggingPayload{}}, CloudLoggingPayload{}},
		{"with-duration", args{CloudLoggingPayload{
			"jsonPayload.duration": "21",
		}}, CloudLoggingPayload{
			"jsonPayload.duration": 21,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixFieldTypes(tt.args.l)
			if diff := cmp.Diff(tt.want, tt.args.l); diff != "" {
				t.Errorf("fixFieldTypes() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
