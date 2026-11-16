// Package smsidentity contains the selected shared SMS identity syntax, not a SIP stack.
package smsidentity

import (
	"strings"

	"crypto/sha256"
	"encoding/binary"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

func HSSID(s string) error {
	if len(s) < 6 || len(s) > 15 {
		return validation.ErrField
	}
	for i := range s {
		if s[i] < '0' || s[i] > '9' {
			return validation.ErrField
		}
	}
	return nil
}

// This is a selected RFC3261 URI grammar, not a SIP stack or URI equivalence
// algorithm. No decoding, IDNA conversion, host folding or percent normalization.
func SIPURI(s string) error {
	if len(s) > 256 || !strings.HasPrefix(s, "sip:") {
		return validation.ErrUnsupported
	}
	address := s[4:]
	at := strings.IndexByte(address, '@')
	if at <= 0 || at == len(address)-1 || strings.LastIndexByte(address, '@') != at {
		return validation.ErrField
	}
	user, host := address[:at], address[at+1:]
	if strings.ContainsAny(host, ":;?[]") {
		return validation.ErrUnsupported
	}
	if avp.Identity(host) != nil {
		return validation.ErrField
	}
	last := host[strings.LastIndexByte(host, '.')+1:]
	if !(last[0] >= 'a' && last[0] <= 'z' || last[0] >= 'A' && last[0] <= 'Z') {
		return validation.ErrUnsupported
	}
	hex := func(b byte) bool { return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F' }
	for i := 0; i < len(user); i++ {
		b := user[i]
		if b == '%' {
			if i+2 >= len(user) || !hex(user[i+1]) || !hex(user[i+2]) {
				return validation.ErrField
			}
			i += 2
			continue
		}
		if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || strings.ContainsRune("-_.!~*'()&=+$,;?/", rune(b)) {
			continue
		}
		return validation.ErrUnsupported
	}
	return nil
}

// URIsDigest binds ordered raw URI bytes with unambiguous lengths. It does not
// normalize URI spelling. Original values stay in the owned request snapshot.
func URIsDigest(origin, destination string) [32]byte {
	return urisDigest("diameterkit/IMS-MO/URIs/v1", origin, destination)
}

// MTURIsDigest uses a separate domain for MT forwarding evidence.
func MTURIsDigest(origin, destination string) [32]byte {
	return urisDigest("diameterkit/IMS-MT/URIs/v1", origin, destination)
}
func urisDigest(domain, origin, destination string) [32]byte {
	h := sha256.New()
	h.Write([]byte(domain))
	var n [4]byte
	for _, s := range []string{origin, destination} {
		binary.BigEndian.PutUint32(n[:], uint32(len(s)))
		h.Write(n[:])
		h.Write([]byte(s))
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}
