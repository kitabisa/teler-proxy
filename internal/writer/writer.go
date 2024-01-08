package writer

import (
	"encoding/json"

	"github.com/charmbracelet/log"
	"github.com/kitabisa/teler-proxy/internal/logger"
	"github.com/kitabisa/teler-proxy/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Writer interface {
	Write(p []byte) (n int, err error)
}

type logWriter struct {
	*log.Logger
	*metrics.Metrics
}

type data map[string]any

func New() *logWriter {
	w := new(logWriter)
	w.Logger = logger.New().WithPrefix("teler-waf")

	return w
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	var d data

	n = len(p)

	err = json.Unmarshal(p, &d)
	if err != nil {
		return 0, err
	}

	w.Logger = w.With("ts", d["ts"])

	err = w.write(d)

	return
}

func (w *logWriter) write(d data) error {
	switch level := d["level"].(string); level {
	case "debug":
		w.writeDebug(d)
	case "info":
		w.writeInfo(d)
	case "warn":
		r, err := json.Marshal(d["request"])
		if err != nil {
			return err
		}

		w.writeWarn(d, r)
	case "error":
		w.writeError(d)
	case "fatal":
		w.writeFatal(d)
	}

	return nil
}

func (w *logWriter) writeDebug(d data) {
	w.Debug(d["msg"])
}

func (w *logWriter) writeInfo(d data) {
	if opt, ok := d["options"].(data); ok {
		w.Info(d["msg"],
			"options", opt,
		)
	}
}

func (w *logWriter) writeWarn(d data, r []byte) {
	var kv []interface{}

	if d["category"] != nil {
		if w.Metrics != nil {
			w.Metrics.Events.With(prometheus.Labels{
				"rule":   d["msg"].(string),
				"threat": d["category"].(string),
			})
		}

		kv = []interface{}{
			"id", d["id"],
			"threat", d["category"],
			"request", string(r),
		}
	}

	w.Warn(d["msg"], kv...)
}

func (w *logWriter) writeError(d data) {
	w.Error(d["msg"],
		"source", d["caller"],
	)
}

func (w *logWriter) writeFatal(d data) {
	w.writeError(d)
}
