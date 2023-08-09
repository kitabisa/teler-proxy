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
	"github.com/kitabisa/teler-proxy/pkg/tunnel"
)

type Runner struct {
	*common.Options
	*fsnotify.Watcher
	*http.Server

	shuttingDown bool
	shutdownLock sync.Mutex
}

func New(opt *common.Options) error {
	reachable := isReachable(opt.Destination, 5*time.Second)
	if !reachable {
		return errDestUnreachable
	}

	run := &Runner{Options: opt}

	if opt.Config.Path != "" {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		if err := watcher.Add(opt.Config.Path); err != nil {
			return err
		}

		run.Watcher = watcher
		defer run.Watcher.Close()

		go func() {
			if err := run.watch(); err != nil {
				opt.Logger.Fatal("Something went wrong", "err", err)
			}
		}()
	}

	dest := buildDest(opt.Destination)

	tun, err := tunnel.NewTunnel(opt.Port, dest, opt.Config.Path, opt.Config.Format)
	if err != nil {
		return err
	}

	logger := log.StandardLog(log.StandardLogOptions{
		ForceLevel: log.ErrorLevel,
	})

	server := &http.Server{
		Addr:     fmt.Sprintf(":%d", opt.Port),
		Handler:  tun,
		ErrorLog: logger,
	}
	run.Server = server

	go func() {
		if err := run.start(); err != nil {
			opt.Logger.Fatal("Something went wrong", "err", err)
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
		"port", r.Options.Port,
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

func (r *Runner) shutdown() error {
	r.shutdownLock.Lock()
	defer r.shutdownLock.Unlock()

	if r.shuttingDown {
		return nil
	}
	r.shuttingDown = true

	r.Options.Logger.Info("Gracefully shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.Server.Shutdown(ctx)
}

func (r *Runner) restart() error {
	r.Options.Logger.Info("Restarting...")

	if err := r.shutdown(); err != nil {
		return err
	}

	return New(r.Options)
}

func (r *Runner) notify(sigCh chan os.Signal) error {
	sig := <-sigCh

	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		return r.shutdown()
	case syscall.SIGHUP:
		return r.restart()
	}

	return nil
}

func (r *Runner) watch() error {
	for {
		select {
		case event := <-r.Watcher.Events:
			if event.Op == 2 {
				r.Options.Logger.Warn("Configuration file has changed", "conf", r.Options.Config.Path)
				return r.restart()
			}
		case err := <-r.Watcher.Errors:
			return err
		}
	}
}
