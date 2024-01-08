package runner

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"net/http"
	"os/signal"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
	"github.com/kitabisa/teler-proxy/common"
	"github.com/kitabisa/teler-proxy/internal/cron"
	"github.com/kitabisa/teler-proxy/internal/writer"
	"github.com/kitabisa/teler-waf"
	"github.com/kitabisa/teler-waf/threat"
)

type Runner struct {
	*common.Options
	*cron.Cron
	*http.Server
	*Prometheus

	shutdown  shutdown
	telerOpts teler.Options
	watcher
}

func New(opt *common.Options) error {
	reachable := isReachable(opt.Destination, 5*time.Second)
	if !reachable {
		return errDestUnreachable
	}

	run := &Runner{Options: opt}

	if opt.Config.Path != "" {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		if err := w.Add(opt.Config.Path); err != nil {
			return err
		}

		defer w.Close()
		run.watcher.config = w
	} else {
		run.watcher.config = new(fsnotify.Watcher)
	}

	if opt.Port.Metrics > 0 {
		if err := run.initMetrics(); err != nil {
			opt.Logger.Error(errInitMetrics, "err", err)
		}
	}

	dest := buildDest(opt.Destination)
	writer := writer.New()
	if run.Prometheus.Metrics != nil {
		writer.Metrics = run.Prometheus.Metrics
	}

	tun, err := run.createTunnel(dest, writer)
	if err != nil {
		return err
	}

	handler := tun.ReverseProxy
	// TODO(dwisisant0):
	//   1. wrap Tunnel server
	//   2. implement Prometheus.Registry in single metrics route
	//   3. implement promMiddleware
	// if run.Prometheus.Metrics != nil {
	// 	handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	// 		promMiddleware(tun).ServeHTTP(w, req)
	// 	})
	// }

	logger := log.StandardLog(log.StandardLogOptions{
		ForceLevel: log.ErrorLevel,
	})

	server := &http.Server{
		Addr:     fmt.Sprintf(":%d", opt.Port.Server),
		Handler:  handler,
		ErrorLog: logger,
	}

	run.Server = server
	run.telerOpts = tun.Options
	run.shutdown.Once = new(sync.Once)

	if run.shouldCron() && run.Cron == nil {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		ds, err := threat.Location()
		if err != nil {
			return err
		}

		if err := w.Add(ds); err != nil {
			return err
		}

		defer w.Close()
		run.watcher.datasets = w

		if err := run.cron(); err != nil {
			opt.Logger.Fatal(errSomething, "err", err)
		}
	} else {
		run.watcher.datasets = new(fsnotify.Watcher)
	}

	go func() {
		if err := run.watch(); err != nil {
			opt.Logger.Fatal(errSomething, "err", err)
		}
	}()

	go func() {
		if err := run.start(); err != nil {
			opt.Logger.Fatal(errSomething, "err", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	return run.notify(sig)
}

func (r *Runner) start() error {
	var err error

	cert := r.Options.TLS.CertPath
	key := r.Options.TLS.KeyPath
	tls := (cert != "" && key != "")

	r.Options.Logger.Info(
		"Server started!",
		"port", r.Options.Port.Server,
		"tls", tls,
		"pid", os.Getpid(),
	)

	if tls {
		err = r.Server.ListenAndServeTLS(cert, key)
	} else {
		err = r.Server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (r *Runner) stop() error {
	r.shutdown.Do(func() {
		r.Options.Logger.Info("Gracefully shutdown...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		r.shutdown.err = r.Server.Shutdown(ctx)
	})

	return r.shutdown.err
}

func (r *Runner) restart() error {
	r.Options.Logger.Info("Restarting...")

	if err := r.stop(); err != nil {
		return err
	}

	return New(r.Options)
}

func (r *Runner) notify(sigCh chan os.Signal) error {
	sig := <-sigCh

	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		return r.stop()
	case syscall.SIGHUP:
		return r.restart()
	}

	return nil
}

func (r *Runner) watch() error {
	for {
		select {
		case event := <-r.watcher.config.Events:
			if event.Op.Has(fsnotify.Write) {
				r.Options.Logger.Warn("Configuration file has changed", "conf", r.Options.Config.Path)
				return r.restart()
			}
		case event := <-r.watcher.datasets.Events:
			if event.Op.Has(fsnotify.Write) || event.Op.Has(fsnotify.Remove) {
				r.Options.Logger.Warn("Threat datasets has updated", "event", event.Op)
				return r.restart()
			}
		case err := <-r.watcher.config.Errors:
			return err
		case err := <-r.watcher.datasets.Errors:
			return err
		}
	}
}

func (r *Runner) cron() error {
	c, err := cron.New()
	if err != nil {
		return err
	}

	r.Cron = c
	c.Scheduler.StartAsync()

	return nil
}
