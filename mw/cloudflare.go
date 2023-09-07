package mw

import (
	"net"
	"net/http"

	"github.com/cesbo/go-router"
)

// CloudFlare IP list
var cloudFlareIPs = []string{
	// https://www.cloudflare.com/ips-v4
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"172.64.0.0/13",
	"131.0.72.0/22",

	// https://www.cloudflare.com/ips-v6
	"2400:cb00::/32",
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
}

// cached network list
var cloudFlareNetworks []*net.IPNet

// init function parses IPs and creates networks
func init() {
	for _, ip := range cloudFlareIPs {
		_, network, err := net.ParseCIDR(ip)
		if err != nil {
			panic(err)
		}
		cloudFlareNetworks = append(cloudFlareNetworks, network)
	}
}

// checkIP function checks if ip in cloudFlareNetworks
func checkIP(addr string) bool {
	ip := net.ParseIP(addr)

	for _, network := range cloudFlareNetworks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// CloudFlareMW is a middleware that checks for a CF-Connecting-IP header
// If request from CloudFlare network then request.RemoteAddr will be set to CF-Connecting-IP
func CloudFlareMW() router.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
				addr, _, _ := net.SplitHostPort(r.RemoteAddr)
				if checkIP(addr) {
					r.RemoteAddr = ip + ":0"
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
