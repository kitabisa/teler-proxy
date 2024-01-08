package common

var (
  // App name
  App = "teler-proxy"
  // Version of teler-proxy itself
  Version = ""
  // Banner of teler-proxy
  Banner = `
    __      __       
   / /____ / /__ ____
  / __/ -_) / -_) __/
  \__/\__/_/\__/_/ proxy ` + Version
  // Usage of teler-proxy
  Usage = `
Usage:
  teler-proxy -d <ADDR>:<PORT> [OPTIONS...]

Options:
  -p, --port <PORT>            Set the local port to listen on (default: 1337)
  -d, --dest <ADDR>:<PORT>     Set the destination address for forwarding requests
  -c, --conf <FILE>            Specify the path to the teler WAF configuration file
  -f, --format <FORMAT>        Specify the configuration file format (json/yaml) (default: yaml)
      --metrics-port <PORT>    Specify the port for exposing metrics data
      --cert <FILE>            Specify the path to the SSL certificate file
      --key <FILE>             Specify the path to the SSL private key file
  -V, --version                Display the current teler-proxy version
  -h, --help                   Display this helps text

Examples:
  teler-proxy -d localhost:80
  teler-proxy -p 8000 -d localhost:80
  teler-proxy -d localhost:80 -c /path/to/teler-waf.conf.yaml
  teler-proxy -d localhost:80 -c /path/to/teler-waf.conf.json -f json

`
)
