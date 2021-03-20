package main

import (
	"encoding/json"
	"strings"
	"sync"
)

// SubscriberStats encodes some values pertaining to messages.
type SubscriberStats struct {
	sync.RWMutex
	Errors          int
	TotalMessages   int
	SampledMessages int
}

// Copy returns a copy of subscriber stats.
func (s *SubscriberStats) Copy() SubscriberStats {
	s.RLock()
	defer s.RUnlock()
	return SubscriberStats{
		Errors:          s.Errors,
		TotalMessages:   s.TotalMessages,
		SampledMessages: s.SampledMessages,
	}
}

func (s *SubscriberStats) String() string {
	s.RLock()
	defer s.RUnlock()
	var b strings.Builder
	json.NewEncoder(&b).Encode(s)
	return b.String()
}

// AddError increments the error count.
func (s *SubscriberStats) AddError(n int) {
	s.Lock()
	defer s.Unlock()
	s.Errors += n
}

// AddTotal increments the total message count.
func (s *SubscriberStats) AddTotal(n int) {
	s.Lock()
	defer s.Unlock()
	s.TotalMessages += n
}

// AddSampled increments the total sampled message count.
func (s *SubscriberStats) AddSampled(n int) {
	s.Lock()
	defer s.Unlock()
	s.SampledMessages += n
}
