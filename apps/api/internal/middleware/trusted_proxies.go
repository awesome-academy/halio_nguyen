package middleware

import (
	"fmt"
	"net"

	"github.com/labstack/echo/v4"
)

// NewIPExtractor builds the X-Forwarded-For extractor that makes
// c.RealIP() (and so the per-IP login limb of D3) see the real client
// behind the Next.js rewrites proxy. Only the explicit CIDRs from
// TRUSTED_PROXY_CIDRS are trusted to set XFF — echo's defaults of trusting
// every loopback, link-local, and RFC1918 address are switched off, since
// "anything on the private network" is broader than "the proxy" and would
// let another internal host spoof the rate limiter. A malformed CIDR is a
// boot error, not a silently-ignored entry (fail closed).
func NewIPExtractor(cidrs []string) (echo.IPExtractor, error) {
	options := []echo.TrustOption{
		echo.TrustLoopback(false),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	}
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("TRUSTED_PROXY_CIDRS: invalid CIDR %q: %w", cidr, err)
		}
		options = append(options, echo.TrustIPRange(ipNet))
	}
	return echo.ExtractIPFromXFFHeader(options...), nil
}
