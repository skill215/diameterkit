// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Attribution and validation order for combined SMS request equality checks.
// SPDX-AI-Date: 2026-11-24

package diameter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"slices"
	"testing"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

func equalityEdit(t *testing.T, raw []byte, path []missingSelector, data []byte) []byte {
	t.Helper()
	// Change only payload bytes. Both histories must preserve identical AVP padding.
	result := bytes.Clone(raw)
	area := result[20:]
	for _, selector := range path {
		d := wire.NewDecoder(area, wire.DefaultLimits())
		for occurrence := 0; ; {
			v, ok, err := d.Next()
			if err != nil || !ok {
				t.Fatalf("missing selector %v: %v", selector, err)
			}
			if v.Key != selector.key {
				continue
			}
			if occurrence == selector.occurrence {
				area = v.Data
				break
			}
			occurrence++
		}
	}
	if len(area) != len(data) {
		t.Fatalf("payload width %d, expected %d", len(data), len(area))
	}
	copy(area, data)
	return result

}

// The same received bytes can be made by changing opposite sides of two valid requests.
func TestCombinedEqualityAttribution(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, relation := range []string{"time", "identity"} {
			t.Run(kind+"/"+relation, func(t *testing.T) {
				req := combinedDuplicateRequest(t, n, co, kind, 13901, 23901)
				m := req.Body.(*combined.CCR)
				original := reqBytes(t, co, req)
				var left, right []missingSelector
				var before, after []byte
				if relation == "time" {
					left, right = lengthPath(avp.Key{Code: 55}), recipientPath(1202)
					before = binary.BigEndian.AppendUint32(nil, uint32(m.EventTimestamp))
					after = make([]byte, 4)
					m.EventTimestamp = 0
					m.ServiceInformation.MMSInformation.Value.SubmissionTime = avp.Some(avp.Time(0))
				} else {
					left, right = subscriptionDataPath(1), equipmentDataPath()
					before = bytes.Clone(m.SubscriptionID[1].SubscriptionIDData)
					after = []byte("001010987654321")
					m.SubscriptionID[1].SubscriptionIDData = bytes.Clone(after)
					m.UserEquipmentInfo.Value.UserEquipmentInfoValue = bytes.Clone(after)
				}
				alternate := reqBytes(t, co, req)
				for _, valid := range [][]byte{original, alternate} {
					if _, err := co.Decode(valid); err != nil {
						t.Fatal(err)
					}
				}
				if bytes.Equal(before, after) {
					t.Fatal("expected distinct inputs")
				}
				for _, pair := range [][2][]byte{
					{equalityEdit(t, original, left, after), equalityEdit(t, alternate, right, before)},
					{equalityEdit(t, original, right, after), equalityEdit(t, alternate, left, before)},
				} {
					if !bytes.Equal(pair[0], pair[1]) {
						t.Fatal("mutation histories produced different wire images")
					}
					_, cause := co.Decode(pair[0])
					if !errors.Is(cause, validation.ErrField) {
						t.Fatal(cause)
					}
					if answer, err := n.rejectFrame(pair[0], 4096, rejectDecode, cause); !errors.Is(err, ErrUnsupported) || len(answer) != 0 {
						t.Fatalf("attributed ambiguous mismatch: %x %v", answer, err)
					}
				}
			})
		}
	}
}

func TestCombinedEqualityOrder(t *testing.T) {
	n, co := newInboundNode(t, combinedConfig(t))
	defer closeDuplicateNode(t, n)
	for _, kind := range []string{"iec", "initial", "term"} {
		req := combinedDuplicateRequest(t, n, co, kind, 13902, 23902)
		raw := reqBytes(t, co, req)
		badTime := equalityEdit(t, raw, recipientPath(1202), make([]byte, 4))
		both := equipmentDataRaw(t, badTime, kind, []byte("different"))
		identity := []avp.Key{{Code: 458}, {Code: 460}}
		stamp := []avp.Key{{VendorID: 10415, Code: 1202}}
		nul := []byte{0}
		for _, tc := range []struct {
			name  string
			raw   []byte
			path  []avp.Key
			quote []byte
		}{
			{"both-relations", both, identity, nil},
			{"equipment-before-client", smsClientRaw(t, both, kind, 2018, nul), identity, nil},
			{"time-before-units", equalityEdit(t, badTime, unitValuePath(kind), unitValueData(kind, 8)), stamp, nil},
			{"client-before-time", smsClientRaw(t, badTime, kind, 2018, nul), nil, smsClientQuote(t, 2018, nul)},
			{"equipment-nul-before-relations", equipmentDataRaw(t, badTime, kind, nul), nil, equipmentDataQuote(t, nul)},
			{"subscription-nul-before-relations", subscriptionDataRaw(t, badTime, 1, nul), nil, subscriptionDataQuote(t, nul)},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				_, cause := co.Decode(tc.raw)
				if !errors.Is(cause, validation.ErrField) {
					t.Fatal(cause)
				}
				if tc.quote == nil {
					var field *validation.FieldError
					if !errors.As(cause, &field) || !slices.Equal(field.Path, tc.path) {
						t.Fatal(cause)
					}
					if ans, err := n.rejectFrame(tc.raw, 4096, rejectDecode, cause); !errors.Is(err, ErrUnsupported) || len(ans) != 0 {
						t.Fatal(err, ans)
					}
					return
				}
				answer, err := n.rejectFrame(tc.raw, 4096, rejectDecode, cause)
				if err != nil {
					t.Fatal(err)
				}
				msg, err := co.Decode(answer)
				if err != nil {
					t.Fatal(err)
				}
				body, ok := msg.Body.(*pdu.ErrorAnswer)
				if !ok || body.ResultCode != 5004 || !body.FailedAVP.Present || !bytes.Equal(body.FailedAVP.Value.Data, tc.quote) {
					t.Fatal(msg.Body)
				}
				if ans, err := n.rejectFrame(tc.raw, len(answer)-1, rejectDecode, cause); err == nil || len(ans) != 0 {
					t.Fatal("short answer budget", err)
				}
			})
		}
	}
}

func TestCombinedEqualityClose(t *testing.T) {
	eachNetwork(t, func(t *testing.T, encrypted bool) {
		for _, kind := range []string{"iec", "initial", "term"} {
			for _, relation := range []string{"time", "identity"} {
				for _, policy := range []struct {
					name      string
					automatic bool
				}{{"default", false}, {"automatic", true}} {
					t.Run(kind+"/"+relation+"/"+policy.name, func(t *testing.T) {
						cfg := combinedConfig(t)
						cfg.Peers[0].InboundCommands[0].FailurePolicy = CloseOnFailure
						if policy.automatic {
							cfg.Peers[0].InboundCommands[0].FailurePolicy = BaseErrorFailures
						}
						n, co := newInboundNode(t, cfg)
						defer closeDuplicateNode(t, n)
						peer, remote := networkConnectNode(t, n, co, encrypted)
						raw := reqBytes(t, co, combinedDuplicateRequest(t, n, co, kind, 13903, 23903))
						if relation == "identity" {
							raw = equipmentDataRaw(t, raw, kind, []byte("different"))
						} else {
							raw = equalityEdit(t, raw, recipientPath(1202), make([]byte, 4))
						}
						sendRejectRaw(t, remote, raw)
						waitFor(t, func() bool { return !peer.Info().Ready })
						if answer, err := readAdapter(ctxTest(t), remote); err == nil {
							t.Fatalf("unexpected reply %x", answer)
						}
					})
				}
			}
		}
	})
}
