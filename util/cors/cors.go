package cors

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

type (
	// IChecker is the the interface that will be used by outside package to interact with this package
	IChecker interface {
		Check(string) bool
	}

	// Checker is the main object for csrf checker package
	checker struct {
		regex *regexp.Regexp
	}
)

var (
	checkerObj      checker
	whitelistOrigin map[string]string
)

// New function will create new csrf checker object
func New() IChecker {

	// if checkerObj.regex == nil {
	// 	reg, _ := regexp.Compile("^(https://)([a-zA-Z-]*)(.tokopedia.com)$|(https://)([a-zA-Z-]*)(.tokopedia.com)([/?])|^(http://)([a-z.-]*)(.ndvl)|^(https://)([a-zA-Z0-9-]*)(staging.tokopedia.com)$|(https://)([a-zA-Z0-9-]*)(staging.tokopedia.com)([/?])|^(http://)(localhost-intools)")
	// 	checkerObj = checker{reg}
	// }

	return &checkerObj
}

// Check for origin and referer header content to validate csrf attack
// will return false if suspected from untrusted source
func (chk *checker) Check(origin string) bool {
	// if chk.regex.MatchString(origin) == true {
	// 	return true
	// }
	return isTokopediaURLCors(origin)

}

// isTokopediaURLCors handles common tokopedia URL except for the one without subdomain.
// This will only handle a URL with protocol.
func isTokopediaURLCors(u string) bool {
	if u == "" {
		return false
	}

	uObject, err := url.Parse(u)
	if err != nil {
		return false
	}

	host, _, err := net.SplitHostPort(uObject.Host)
	if err != nil {
		host = uObject.Host
	}

	if !strings.HasSuffix(host, ".tokopedia.com") &&
		!strings.HasSuffix(host, ".tokopedia.net") &&
		!strings.HasSuffix(host, "devel-go.tkpd") &&
		!strings.HasSuffix(host, ".ndvl") &&
		!strings.HasSuffix(host, ".tokopedia.id") {
		return false
	}

	return true
}

// we can add more origin client that we trusted here
func initWhitelistOrigin() {
	if whitelistOrigin == nil {
		whitelistOrigin = make(map[string]string)
	}
	// staging
	whitelistOrigin["https://staging.tokopedia.com"] = "https://staging.tokopedia.com"
	whitelistOrigin["https://m-staging.tokopedia.com"] = "https://m-staging.tokopedia.com"
	whitelistOrigin["https://internal-staging.tokopedia.com"] = "https://internal-staging.tokopedia.com"
	whitelistOrigin["https://tome-staging.tokopedia.com"] = "https://tome-staging.tokopedia.com"
	whitelistOrigin["https://ims-staging.tokopedia.com"] = "https://ims-staging.tokopedia.com"
	whitelistOrigin["https://seller-staging.tokopedia.com"] = "https://seller-staging.tokopedia.com"

	// prod
	whitelistOrigin["https://www.tokopedia.com"] = "https://www.tokopedia.com"
	whitelistOrigin["https://m.tokopedia.com"] = "https://m.tokopedia.com"
	whitelistOrigin["https://internal.tokopedia.com"] = "https://internal.tokopedia.com"
	whitelistOrigin["https://tome.tokopedia.com"] = "https://tome.tokopedia.com"
	whitelistOrigin["https://ims.tokopedia.com"] = "https://ims.tokopedia.com"
	whitelistOrigin["https://seller.tokopedia.com"] = "https://seller.tokopedia.com"
}

// CheckWhitelistOrigin filter unknown origin
func CheckWhitelistOrigin(origin string) (trustedOrigin string, trusted bool) {
	if whitelistOrigin == nil {
		initWhitelistOrigin()
	}

	if trustedOrigin, trusted = whitelistOrigin[origin]; trusted {
		return
	}

	// empty string and false (untrusted)
	return
}
