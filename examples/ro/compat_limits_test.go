package main

import (
	"context"
	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"testing"
)

func TestCompatOfflineLimits(t *testing.T) {
	cfg, _, e := compatConfigs()
	if e != nil {
		t.Fatal(e)
	}
	n, e := diameter.NewNode(cfg[0])
	if e != nil {
		t.Fatal(e)
	}
	defer func() { n.Close(); n.Wait(context.Background()) }()
	body, e := compatRequest("debit", "test;1")
	if e != nil {
		t.Fatal(e)
	}
	r, e := n.PrepareRoRecoveryRequest(context.Background(), body, compatOptions())
	if e != nil {
		t.Fatal(e)
	}
	c, e := codec.New(cfg[1].Bundle, codec.Config{Limits: cfg[1].Limits})
	if e != nil {
		t.Fatal(e)
	}
	op, ok := cfg[1].Bundle.Lookup(body.CommandKey())
	if !ok {
		t.Fatal("missing command")
	}
	rb, e := op.DecodeBound(1024, cfg[1].Limits)
	if e != nil {
		t.Fatal(e)
	}
	ab, e := c.DecodeBound(pdu.Header{ApplicationID: 4, CommandCode: 272, Flags: pdu.ProxiableFlag}, 1024)
	if e != nil {
		t.Fatal(e)
	}
	reserve := 8*(rb+ab) + 131072
	if reserve <= 32<<20 || reserve > cfg[1].Peers[0].MaxInboundBytes || reserve > cfg[1].Resources.InboundClasses[0].Bytes {
		t.Fatal("combined inbound budget", reserve)
	}
	_, e = c.Decode(r.Frame)
	if e != nil {
		t.Fatal(e)
	}
}
