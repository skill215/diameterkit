// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Bounded Diameter-only URI validation without resolution or dialing.
// SPDX-AI-Date: 2026-11-17

package avp

import (
	"net/netip"
	"strings"
)

// MaxDiameterURIBytes bounds the selected DNS-name, five-digit port and
// Diameter/TCP-or-SCTP URI grammar, including its longest optional suffix.
const MaxDiameterURIBytes = len("aaas://") + 253 + len(":65535;transport=sctp;protocol=diameter")

// DiameterURI validates a selected RFC 6733 DiameterURI subset. It accepts aaa
// and aaas, Identity's ASCII DNS-name subset, ports 1–65535 (at most five digits),
// optional tcp/sctp transport and optional diameter protocol, in that order.
// Tokens are ASCII case-insensitive; the input is never normalized or resolved.
// IP literals, trailing root dots, escapes, userinfo, paths, queries, fragments,
// UDP, other AAA protocols and unknown/repeated/reordered parameters are rejected.
func DiameterURI(s string) error {
	if len(s) == 0 || len(s) > MaxDiameterURIBytes {
		return ErrValue
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 33 || s[i] > 126 {
			return ErrValue
		}
	}
	scheme, rest, ok := strings.Cut(s, "://")
	if !ok || !strings.EqualFold(scheme, "aaa") && !strings.EqualFold(scheme, "aaas") {
		return ErrValue
	}
	authority, params, hasParams := strings.Cut(rest, ";")
	host, port, hasPort := strings.Cut(authority, ":")
	if Identity(host) != nil {
		return ErrValue
	}
	if _, err := netip.ParseAddr(host); err == nil {
		return ErrValue
	}
	if hasPort {
		if len(port) == 0 || len(port) > 5 {
			return ErrValue
		}
		n := 0
		for i := 0; i < len(port); i++ {
			if port[i] < '0' || port[i] > '9' {
				return ErrValue
			}
			n = n*10 + int(port[i]-'0')
		}
		if n == 0 || n > 65535 {
			return ErrValue
		}
	}
	if !hasParams {
		return nil
	}
	first, last, more := strings.Cut(params, ";")
	transport := strings.EqualFold(first, "transport=tcp") || strings.EqualFold(first, "transport=sctp")
	if more {
		if !transport || !strings.EqualFold(last, "protocol=diameter") {
			return ErrValue
		}
	} else if !transport && !strings.EqualFold(first, "protocol=diameter") {
		return ErrValue
	}
	return nil
}
