package sctp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
	"github.com/skill215/diameterkit/wire"
	upstream "github.com/skill215/sctp"
)

func modeFrame(session string) []byte {
	b := make([]byte, 20+8+len(session)+4)
	w := wire.NewEncoder(b[20:], wire.DefaultLimits())
	if session != "" {
		w.Bytes(avp.Key{Code: 263}, 64, []byte(session))
	}
	b = b[:20+w.Len()]
	wire.PutHeader(b, pdu.Header{Flags: pdu.RequestFlag, ApplicationID: 16777312, CommandCode: 8388647, HopByHop: 11, EndToEnd: 22}, len(b))
	return b
}

func TestSCTPModeHashAndPolicy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StreamMode = OrderedMultiStream
	for _, session := range []string{"", "a", "ctf.example;1;2"} {
		b := modeFrame(session)
		h, e := streamHash(b)
		if e != nil {
			t.Fatal(e)
		}
		// Independent fixed FNV vectors: domains + exact session or header key.
		want := map[string]uint64{"": 0xa295ec0cae4dff3f, "a": 0x82f4307b4e8c4d7, "ctf.example;1;2": 0x175e72c7bd76f7ee}[session]
		if h != want {
			t.Fatalf("%q hash=%#x want=%#x", session, h, want)
		}
		for _, count := range []uint16{1, 2, 3, 16, 65535} {
			i, e := sendInfo(cfg, count, b, false)
			if e != nil || i.Flags != 0 || i.PPID != 46 || i.SID >= count || count > 1 && i.SID == 0 {
				t.Fatal(i, e)
			}
			i2, e := sendInfo(cfg, count, b, true)
			if e != nil || i2.SID != 0 || i2.Flags != 0 {
				t.Fatal("barrier", i2, e)
			}
		}
		b[4] = 0
		binary.BigEndian.PutUint32(b[12:16], 99)
		h2, e := streamHash(b)
		if e != nil || h2 != h {
			t.Fatal("answer/HBH key drift", h2, e)
		}
		b[4] = pdu.RequestFlag | pdu.RetransmittedFlag
		h2, e = streamHash(b)
		if e != nil || h2 != h {
			t.Fatal("retransmission key drift", h2, e)
		}
		if session != "" {
			binary.BigEndian.PutUint32(b[8:12], 4)
			wire.Put24(b[5:8], 272)
			binary.BigEndian.PutUint32(b[16:20], 77)
			h2, e = streamHash(b)
			if e != nil || h2 != h {
				t.Fatal("same session key drift", h2, e)
			}
		}
		if n := testing.AllocsPerRun(100, func() {
			_, err := streamHash(b)
			if err != nil {
				panic(err)
			}
		}); n != 0 {
			t.Fatal("payload selection allocated", n)
		}
	}
	for _, mode := range []StreamMode{UnorderedStream0, OrderedMultiStream} {
		for _, policy := range []PPIDPolicy{PPID46, PPID0Compatibility} {
			cfg.StreamMode, cfg.PPIDPolicy = mode, policy
			for _, code := range []uint32{257, 280, 282} {
				b := modeFrame("s")
				binary.BigEndian.PutUint32(b[8:12], 0)
				wire.Put24(b[5:8], code)
				i, e := sendInfo(cfg, 3, b, false)
				if e != nil || i.SID != 0 || i.Flags != 0 || i.PPID != map[PPIDPolicy]uint32{PPID46: 46, PPID0Compatibility: 0}[policy] {
					t.Fatal("control", i, e)
				}
			}
			b := modeFrame("s")
			binary.BigEndian.PutUint32(b[8:12], 0)
			wire.Put24(b[5:8], 271) // SCAP1 is application, not Base control.
			i, e := sendInfo(cfg, 3, b, false)
			if e != nil || mode == OrderedMultiStream && i.SID == 0 || mode == UnorderedStream0 && (i.SID != 0 || i.Flags != upstream.SCTP_UNORDERED) {
				t.Fatal("App0 accounting", i, e)
			}
		}
	}
}

func TestSCTPModeMalformedSelectionAndNoState(t *testing.T) {
	valid := modeFrame("session")
	for _, name := range []string{"duplicate", "empty", "truncated", "badlength", "version"} {
		t.Run(name, func(t *testing.T) {
			b := append([]byte(nil), valid...)
			switch name {
			case "duplicate":
				b = append(b, b[20:]...)
				wire.Put24(b[1:4], uint32(len(b)))
			case "empty":
				b = b[:28]
				wire.Put24(b[1:4], 28)
				wire.Put24(b[25:28], 8)
			case "truncated":
				b = b[:len(b)-1]
			case "badlength":
				wire.Put24(b[25:28], 7)
			case "version":
				b[0] = 2
			}
			if _, e := streamHash(b); !errors.Is(e, transport.ErrFrameState) {
				t.Fatal(e)
			}
		})
	}
	h, _ := streamHash(valid)
	valid[len(valid)-1] = 1 // Diameter padding is ignored, not part of Session-Id.
	if got, e := streamHash(valid); e != nil || got != h {
		t.Fatal("padding changed key", got, e)
	}
	// Only a top-level vendor0 Session-Id supplies the session key. Neither
	// a vendor extension nor an opaque grouped child may override it.
	for _, grouped := range []bool{false, true} {
		b := make([]byte, 128)
		copy(b, valid)
		w := wire.NewEncoder(b[len(valid):], wire.DefaultLimits())
		if grouped {
			mark := w.BeginGroup(avp.Key{Code: 284}, 64)
			w.Bytes(avp.Key{Code: 263}, 64, []byte("other"))
			w.EndGroup(mark)
		} else {
			w.Bytes(avp.Key{VendorID: 10415, Code: 263}, 192, []byte("other"))
		}
		b = b[:len(valid)+w.Len()]
		wire.Put24(b[1:4], uint32(len(b)))
		if got, e := streamHash(b); e != nil || got != h {
			t.Fatal("wrong parent/vendor changed key", got, e)
		}
	}
	// Distinct sessions exercise more than one lane without storing a session table.
	seen := map[uint16]bool{}
	cfg := DefaultConfig()
	cfg.StreamMode = OrderedMultiStream
	for i := 0; i < 32; i++ {
		s, e := sendInfo(cfg, 4, modeFrame(fmt.Sprint(i)), false)
		if e != nil {
			t.Fatal(e)
		}
		seen[s.SID] = true
	}
	if len(seen) != 3 {
		t.Fatal(seen)
	}
	if _, e := sendInfo(cfg, 0, valid, false); !errors.Is(e, ErrConfig) {
		t.Fatal(e)
	}
}

func FuzzSCTPModeSelection(f *testing.F) {
	f.Add(modeFrame("session"), uint16(3))
	f.Add(modeFrame(""), uint16(1))
	f.Fuzz(func(t *testing.T, b []byte, n uint16) {
		if len(b) > 1<<20 {
			return
		}
		h, e := streamHash(b)
		if e != nil {
			return
		}
		cfg := DefaultConfig()
		cfg.StreamMode = OrderedMultiStream
		i, e := sendInfo(cfg, n, b, false)
		if n == 0 {
			if !errors.Is(e, ErrConfig) {
				t.Fatal(e)
			}
			return
		}
		if e != nil || i.SID >= n || i.Flags != 0 {
			t.Fatal(h, i, e)
		}
	})
}
