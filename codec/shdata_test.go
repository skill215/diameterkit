package codec_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/validation"
)

func TestShDataExplicitInterpretationAndCanonical(t *testing.T) {
	c := shCodec(t, codec.Config{})
	for _, tc := range []struct {
		frame, xml string
		use        shdata.Use
	}{
		{"v-sh-pur-01", "x-sh-reg-01", shdata.SMSUpdate},
		{"v-sh-uda-01", "x-sh-pull-01", shdata.SMSPull},
		{"v-sh-pnr-01", "x-sh-mme-01", shdata.ReachabilityNotification},
	} {
		t.Run(tc.frame, func(t *testing.T) {
			m := decoded(t, c, tc.frame)
			var raw []byte
			switch b := m.Body.(type) {
			case *sh.PUR:
				raw = b.UserData
			case *sh.UDA:
				raw = b.UserData.Value
			case *sh.PNR:
				raw = b.UserData
			}
			expected, e := os.ReadFile("../testdata/shdata/" + tc.xml + ".xml")
			if e != nil || !bytes.Equal(raw, expected) {
				t.Fatal("frozen AVP/XML pair", e)
			}
			before := encoded(t, c, m)
			d, e := shdata.Parse(raw, shdata.DefaultLimits())
			if e != nil {
				t.Fatal(e)
			}
			report, e := d.Check(tc.use)
			if e != nil || !report.Complete {
				t.Fatal(report, e)
			}
			if !bytes.Equal(before, encoded(t, c, m)) {
				t.Fatal("helper altered original message")
			}
		})
	}
	// Semantically equal deregistration XML remains different duplicate content.
	m := decoded(t, c, "v-sh-pur-01")
	body := m.Body.(*sh.PUR)
	body.UserData, _ = shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: ""}, shdata.DefaultLimits())
	first, e := c.Canonical(encoded(t, c, m), 4096)
	if e != nil {
		t.Fatal(e)
	}
	a, e := shdata.Parse(body.UserData, shdata.DefaultLimits())
	if e != nil {
		t.Fatal(e)
	}
	body.UserData = []byte(strings.Replace(string(body.UserData), "<IP-SM-GW-Number></IP-SM-GW-Number>", "<IP-SM-GW-Number/>", 1))
	b, e := shdata.Parse(body.UserData, shdata.DefaultLimits())
	if e != nil || a.Data() != b.Data() {
		t.Fatal("expected same interpreted data", e)
	}
	second, e := c.Canonical(encoded(t, c, m), 4096)
	if e != nil || bytes.Equal(first, second) {
		t.Fatal("XML interpretation replaced raw identity", e)
	}
	// XML helper limits/errors cannot silently change the codec's scoped report.
	body.UserData = []byte("<broken")
	if r := c.ValidateContext(m, validation.Context{}); !r.Complete {
		t.Fatal(r)
	}
	if _, e = shdata.Parse(body.UserData, shdata.DefaultLimits()); !errors.Is(e, shdata.ErrXML) {
		t.Fatal(e)
	}
	if e = c.Validate(m); e != nil {
		t.Fatal("helper changed AVP acceptance", e)
	}
}
