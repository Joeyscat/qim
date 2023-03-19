package qim

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

var cidrs []*net.IPNet

func init() {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link-local block
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link-local address IPv6
	}

	cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}
}

func isPrivateAddress(address string) (bool, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		return false, errors.New("invalid IP address")
	}
	for _, cidr := range cidrs {
		if cidr.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

// FromRequest returns the real IP address of the client.
// It parses the X-Forwarded-For header or the X-Real-Ip header
// if they are present in the request.
// Otherwise it returns the request.RemoteAddr.
// If the request is nil then it returns an empty string.
func FromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	// X-Forwarded-For: client, proxy1, proxy2
	// X-Real-Ip: client
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		for _, addr := range strings.Split(xForwardedFor, ",") {
			addr = strings.TrimSpace(addr)
			isPrivate, err := isPrivateAddress(addr)
			if !isPrivate && err == nil {
				return addr
			}
		}
	} else if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}

	// Parse the IP:port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
