package csrf

import (
	"net"
	"net/url"
	"strings"
)

type (
	IChecker interface {
		Check(string, string) bool
	}
	checker struct{}
)

var checkerObj checker

func New() IChecker {
	return &checker{}
}

func (c *checker) Check(origin, referer string) bool {
	if isGbtURL(origin) || isGbtURL(referer) {
		return true
	}
	return false
}

func isGbtURL(u string) bool {
	if u == "" {
		return false
	}

	uObject, err := url.Parse(u)
	if err != nil {
		return false
	}

	host, _, err := net.SplitHostPort(uObject.Host)
	if err != nil {
		return false
	}

	if host != "gbt.com" &&
		!strings.HasSuffix(host, ".gbt.com") &&
		!strings.HasSuffix(host, ".gbt.net") &&
		!strings.HasSuffix(host, ".gbt.id") {
		return false
	}

	return true
}
