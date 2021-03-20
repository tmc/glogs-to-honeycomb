package main

// CloudLoggingPayload is a Google Cloud Logging payload.
type CloudLoggingPayload map[string]interface{}

// SamplePolicies encodes static sampling rules.
var SamplePolicies = map[string]map[string]uint{
	"labels.k8s-pod/service_istio_io/canonical-name": {
		"opentelemetry-collector": 1000,
	},
}

// ApplySamplingPolicy returns the sample rate for a payload.
// It returns the sample rate for the event.
func ApplySamplingPolicy(l CloudLoggingPayload, defaultSampleRate int) uint {
	if isLikelyError(l) {
		return 1
	}
	// TODO: respect event-specified sample rate.
	for policyKey, policy := range SamplePolicies {
		val, ok := l[policyKey].(string)
		if ok {
			sampleRate, ok := policy[val]
			if ok {
				return sampleRate
			}
		}
	}
	return uint(defaultSampleRate)
}

func isLikelyError(l CloudLoggingPayload) bool {
	// first check response flags
	responseFlags, ok := l["jsonPayload.response_flags"].(string)
	if ok && responseFlags != "" && responseFlags != "-" {
		return true
	}

	// next check for response codes
	responseCode, ok := l["jsonPayload.response_code"].(int)
	if !ok {
		return false
	}
	return responseCode >= 500
}
