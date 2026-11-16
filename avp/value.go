// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package avp defines Diameter values independent of applications and transports.
package avp

import (
	"encoding/binary"
	"errors"
	"net/netip"
	"strings"
	"unicode/utf8"
)

type Key struct{ VendorID, Code uint32 }

const (
	VendorFlag    byte = 0x80
	MandatoryFlag byte = 0x40
	ProtectedFlag byte = 0x20
)

// Optional distinguishes absence from an explicitly supplied zero value.
type Optional[T any] struct {
	Value   T
	Present bool
}

func Some[T any](v T) Optional[T] { return Optional[T]{v, true} }

// CloneString avoids retaining a large caller-owned backing allocation through
// an otherwise small substring in a bounded snapshot.
func CloneString(s string) string { return strings.Clone(s) }

// Raw retains an uninterpreted payload. Unknown payloads are never guessed to be Grouped.
type Raw struct {
	Key   Key
	Flags byte
	Data  []byte
}

func (r Raw) Clone() Raw { r.Data = append([]byte(nil), r.Data...); return r }

var ErrValue = errors.New("invalid AVP value")

// Identity validates the ASCII DNS-name subset used by this release.
// It does not resolve names, normalize Unicode, or authenticate an identity.
func Identity(s string) error {
	if len(s) == 0 || len(s) > 253 {
		return ErrValue
	}
	for label := range strings.SplitSeq(s, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return ErrValue
		}
		for _, c := range []byte(label) {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return ErrValue
			}
		}
	}
	return nil
}
func UTF8(s string) error {
	if !utf8.ValidString(s) || strings.IndexByte(s, 0) >= 0 {
		return ErrValue
	}
	return nil
}

// EncodeTBCD encodes decimal digits without TON/NPI. It does not encode SME-Address.
func EncodeTBCD(digits string) ([]byte, error) {
	if len(digits) == 0 || len(digits) > 15 {
		return nil, ErrValue
	}
	b := make([]byte, (len(digits)+1)/2)
	for i, c := range []byte(digits) {
		if c < '0' || c > '9' {
			return nil, ErrValue
		}
		if i%2 == 0 {
			b[i/2] = c - '0'
		} else {
			b[i/2] |= (c - '0') << 4
		}
	}
	if len(digits)%2 != 0 {
		b[len(b)-1] |= 0xf0
	}
	return b, nil
}
func CheckTBCD(b []byte) error {
	if len(b) == 0 || len(b) > 8 {
		return ErrValue
	}
	n := len(b) * 2
	for i, v := range b {
		if v&15 > 9 {
			return ErrValue
		}
		if v>>4 == 15 && i == len(b)-1 {
			n--
			continue
		}
		if v>>4 > 9 {
			return ErrValue
		}
	}
	if n > 15 {
		return ErrValue
	}
	return nil
}
func DecodeTBCD(b []byte) (string, error) {
	if err := CheckTBCD(b); err != nil {
		return "", err
	}
	out := make([]byte, 0, len(b)*2)
	for _, v := range b {
		out = append(out, '0'+v&15)
		if v>>4 != 15 {
			out = append(out, '0'+(v>>4))
		}
	}
	return string(out), nil
}

// Address retains an explicit address family; non-IP families remain opaque.
type Address struct {
	Family uint16
	Data   []byte
}

func IPAddress(ip netip.Addr) (Address, error) {
	if !ip.IsValid() || ip.Zone() != "" {
		return Address{}, ErrValue
	}
	ip = ip.Unmap()
	f := uint16(2)
	if ip.Is4() {
		f = 1
	}
	return Address{f, ip.AsSlice()}, nil
}
func (a Address) Validate() error {
	if a.Family == 0 || len(a.Data) == 0 || a.Family == 1 && len(a.Data) != 4 || a.Family == 2 && len(a.Data) != 16 {
		return ErrValue
	}
	return nil
}

// ParseAddress borrows b. Codec entry points establish ownership before invoking it.
func ParseAddress(b []byte) (Address, error) {
	if len(b) < 2 {
		return Address{}, ErrValue
	}
	a := Address{binary.BigEndian.Uint16(b), b[2:len(b):len(b)]}
	return a, a.Validate()
}

// Time is the four-octet NTP seconds field used by Diameter Time AVPs. Zero is
// a valid present value (including the 2036 rollover), not an absent timestamp.
// An era must be supplied by the application when interpreting these wire bits;
// the codec never guesses an era, compares wall time or schedules a callback.
type Time uint32
