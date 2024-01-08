package runner

import (
	"flag"
	"os"

	"github.com/kitabisa/teler-proxy/common"
	"github.com/kitabisa/teler-proxy/internal/logger"
)

func ParseOptions() *common.Options {
	opt := new(common.Options)
	cfg := new(common.Config)
	tls := new(common.TLS)
	port := new(common.Port)

	opt.Logger = logger.New()

	flag.IntVar(&port.Server, "p", 1337, "")
	flag.IntVar(&port.Server, "port", 1337, "")

	flag.StringVar(&opt.Destination, "d", "", "")
	flag.StringVar(&opt.Destination, "dest", "", "")

	flag.StringVar(&cfg.Path, "c", "", "")
	flag.StringVar(&cfg.Path, "conf", "", "")

	flag.IntVar(&port.Metrics, "metrics-port", 0, "")

	flag.StringVar(&cfg.Format, "f", "yaml", "")
	flag.StringVar(&cfg.Format, "format", "yaml", "")

	flag.StringVar(&tls.CertPath, "cert", "", "")
	flag.StringVar(&tls.KeyPath, "key", "", "")

	flag.BoolVar(&version, "V", false, "")
	flag.BoolVar(&version, "version", false, "")

	flag.Usage = func() {
		common.PrintBanner()
		common.PrintUsage()
	}
	flag.Parse()

	if version {
		common.PrintVersion()
		os.Exit(1)
	}

	if opt.Destination == "" {
		common.PrintBanner()
		opt.Logger.Fatal("Something went wrong", "err", "missing destination address")
	}

	common.PrintBanner()

	opt.Config = cfg
	opt.TLS = tls
	opt.Port = port

	return opt
}
