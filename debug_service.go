package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"sync"

	"github.com/felixge/fgprof"
)

// DebugService provides a mechanism to expose process internals over http.
type DebugService struct {
	mux *http.ServeMux

	mutex   sync.RWMutex // protects the following fields.
	urls    []string
	expVars map[string]interface{}
}

// NewDebugService creates a new DebugService.
func NewDebugService() (*DebugService, error) {
	s := &DebugService{
		mux:     http.NewServeMux(),
		expVars: make(map[string]interface{}),
	}
	// Add to the mux but don't add an index entry.
	s.mux.HandleFunc("/", s.indexHandler)

	s.HandleFunc("/debug/pprof/", pprof.Index)
	s.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.HandleFunc("/debug/pprof/trace", pprof.Trace)
	s.HandleFunc("/debug/vars", s.expvarHandler)
	s.Handle("/debug/fgprof/profile", fgprof.Handler())
	s.Publish("cmdline", os.Args)
	s.Publish("memstats", Func(memstats))

	return s, nil
}

func (s *DebugService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Handle registers an additional debug endpoint.
func (s *DebugService) Handle(pattern string, handler http.Handler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.urls = append(s.urls, pattern)
	s.mux.Handle(pattern, handler)
}

// HandleFunc registers an additional debug endpoint.
func (s *DebugService) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.urls = append(s.urls, pattern)
	s.mux.HandleFunc(pattern, handler)
}

// Publish an expvar at /debug/vars, possibly using Func
func (s *DebugService) Publish(name string, v interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, existing := s.expVars[name]; existing {
		log.Panicln("Reuse of exported var name:", name)
	}
	s.expVars[name] = v
}

func (s *DebugService) indexHandler(w http.ResponseWriter, req *http.Request) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if err := indexTmpl.Execute(w, s.urls); err != nil {
		log.Println("error rendering debug index:", err)
	}
}

var indexTmpl = template.Must(template.New("index").Parse(`
<html>
<head>
<title>Debug Index</title>
</head>
<body>
<h2>Index</h2>
<table>
{{range .}}
<tr><td><a href="{{.}}?debug=1">{{.}}</a>
{{end}}
</table>
</body>
</html>
`))

func (s *DebugService) expvarHandler(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	values := make(map[string]interface{}, len(s.expVars))
	for k, v := range s.expVars {
		if f, ok := v.(Func); ok {
			v = f()
		}
		values[k] = v
	}
	b, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		log.Println("error encoding expvars:", err)
	}
	w.Write(b)
}

func memstats() interface{} {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return *stats
}

type Func func() interface{}
