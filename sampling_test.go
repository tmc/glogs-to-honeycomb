package main

import "testing"

func TestApplySamplingPolicy(t *testing.T) {
	type args struct {
		l                 CloudLoggingPayload
		defaultSampleRate int
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{"empty", args{CloudLoggingPayload{}, 10}, 10},
		{"otel-trace", args{CloudLoggingPayload{
			"labels.k8s-pod/service_istio_io/canonical-name": "opentelemetry-collector",
		}, 10}, 1000},
		{"other-service", args{CloudLoggingPayload{
			"labels.k8s-pod/service_istio_io/canonical-name": "foo-service",
		}, 10}, 10},
		{"other-service-with-500", args{CloudLoggingPayload{
			"labels.k8s-pod/service_istio_io/canonical-name": "foo-service",
			"jsonPayload.response_code":                      500,
		}, 10}, 1},
		{"other-service-with-UC", args{CloudLoggingPayload{
			"labels.k8s-pod/service_istio_io/canonical-name": "foo-service",
			"jsonPayload.response_code":                      504,
			"jsonPayload.response_flags":                     "UC",
		}, 10}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplySamplingPolicy(tt.args.l, tt.args.defaultSampleRate); got != tt.want {
				t.Errorf("ApplySamplingPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}
