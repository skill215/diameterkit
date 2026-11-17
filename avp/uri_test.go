// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Diameter URI subset, boundaries and bounded fuzz acceptance.
// SPDX-AI-Date: 2026-11-17

package avp_test

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"strings"
	"testing"
)

func TestDiameterURI(t *testing.T) {
	for _, s := range []string{"aaa://host.example", "aaas://host.example", "AAA://Host.Example:00001;TRANSPORT=TCP;PROTOCOL=DIAMETER", "aaa://host.example:65535", "aaa://host.example;transport=sctp", "aaas://host.example;protocol=diameter", "aaa://host.example:3868;transport=tcp;protocol=diameter", "aaas://host.example:5658;transport=sctp;protocol=diameter", "aaa://xn--bcher-kva.example"} {
		t.Run(s, func(t *testing.T) {
			if e := avp.DiameterURI(s); e != nil {
				t.Fatal(e)
			}
		})
	}
	for _, s := range []string{"", "aaa://", "http://host.example", "aaa:/host.example", "aaa://host.example/", "aaa://host.example?q=1", "aaa://host.example#x", "aaa://user@host.example", "aaa://host%2eexample", "aaa://host.example.", "aaa://bad_name.example", "aaa://-host.example", "aaa://a..b", "aaa://host.exämple", "aaa://127.0.0.1", "aaa://[::1]", "aaa://::1", "aaa://host.example:", "aaa://host.example:0", "aaa://host.example:65536", "aaa://host.example:100000", "aaa://host.example:-1", "aaa://host.example:+1", "aaa://host.example:12x", "aaa://host.example:1:2", "aaa://host.example;", "aaa://host.example;transport=udp", "aaa://host.example;protocol=radius", "aaa://host.example;protocol=tacacs+", "aaa://host.example;transport=quic", "aaa://host.example;transport=tcp;transport=sctp", "aaa://host.example;protocol=diameter;transport=tcp", "aaa://host.example;protocol=diameter;protocol=diameter", "aaa://host.example;transport=tcp;protocol=diameter;x=1", "aaa://host.example;unknown=x", "aaa://host.example;transport=ſctp", "aaa://host.example\x00", "aaa://host.example\n", " aaa://host.example", "aaa://host.example ", "aaa://" + strings.Repeat("a", 64) + ".example"} {
		t.Run(s, func(t *testing.T) {
			if !errors.Is(avp.DiameterURI(s), avp.ErrValue) {
				t.Fatal("accepted URI", s)
			}
		})
	}
}

func TestDiameterURIMaximum(t *testing.T) {
	host := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)
	s := "aaas://" + host + ":65535;transport=sctp;protocol=diameter"
	if len(s) != avp.MaxDiameterURIBytes || avp.DiameterURI(s) != nil {
		t.Fatal("maximum", len(s), avp.MaxDiameterURIBytes)
	}
	if avp.DiameterURI(s+"x") == nil || avp.DiameterURI("aaa://"+host+"x") == nil {
		t.Fatal("size bound")
	}
}

func FuzzDiameterURI(f *testing.F) {
	for _, s := range []string{"aaa://host.example", "aaas://host.example:5658;transport=sctp;protocol=diameter", "aaa://host.example;transport=udp", "aaa://[::1]", ""} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if avp.DiameterURI(s) != nil {
			return
		}
		if len(s) == 0 || len(s) > avp.MaxDiameterURIBytes {
			t.Fatal("unbounded URI")
		}
		for _, b := range []byte(s) {
			if b < 33 || b > 126 {
				t.Fatal("non-ASCII URI")
			}
		}
		for _, suffix := range []string{"\x00", "/", "?x=1", "#x", ";unexpected=x"} {
			if avp.DiameterURI(s+suffix) == nil {
				t.Fatal("ignored suffix")
			}
		}
	})
}
