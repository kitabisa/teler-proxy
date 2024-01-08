package runner

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"net/url"

	"github.com/kitabisa/teler-proxy/pkg/tunnel"
)

func parseURL(dest string) (*url.URL, error) {
	if !strings.HasPrefix(dest, "http://") && !strings.HasPrefix(dest, "https://") {
		dest = "http://" + dest
	}

	destURL, err := url.Parse(dest)
	if err != nil {
		return nil, err
	}

	return destURL, nil
}

func cleanURL(inputURL string) string {
	parsedURL, err := parseURL(inputURL)
	if err != nil {
		// Return the input URL as-is if parsing fails
		return inputURL
	}

	return parsedURL.Host
}

func buildDest(dest string) string {
	parsedURL, err := parseURL(dest)
	if err != nil {
		// Return the input URL as-is if parsing fails
		return dest
	}

	return fmt.Sprint(parsedURL.Scheme, "://", parsedURL.Host)
}

func isReachable(inputURL string, timeout time.Duration) bool {
	host := cleanURL(inputURL)

	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

func (r *Runner) shouldCron() bool {
	if r.Options.Config.Path == "" {
		return true
	}

	opt := r.telerOpts

	if !opt.InMemory && !opt.NoUpdateCheck {
		return true
	}

	return false
}

func (r *Runner) createTunnel(dest string, writer io.Writer) (*tunnel.Tunnel, error) {
	opt := r.Options

	return tunnel.NewTunnel(
		opt.Port.Server,
		dest, opt.Config.Path,
		opt.Config.Format,
		writer,
	)
}
