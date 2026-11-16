// Command codec demonstrates an offline SRR/SRA pair; it opens no sockets.
package main

import (
	"fmt"
	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
)

func run() error {
	profile, err := s6c.Profile()
	if err != nil {
		return err
	}
	bundle, err := schema.Compile(profile)
	if err != nil {
		return err
	}
	c, err := codec.New(bundle, codec.Config{})
	if err != nil {
		return err
	}
	msisdn, err := avp.EncodeTBCD("12025550123")
	if err != nil {
		return err
	}
	sc, err := avp.EncodeTBCD("12025550100")
	if err != nil {
		return err
	}
	body := &s6c.SRR{SessionID: "gmsc.example;17;1", AuthSessionState: 1, MSISDN: avp.Some(msisdn), SCAddress: sc, SMRPMTI: avp.Some(uint32(0))}
	request, err := c.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: "gmsc.example", OriginRealm: "example.net", DestinationHost: "hss.example", DestinationRealm: "example.net", HopByHop: 0x51525354, EndToEnd: 0x61626364})
	if err != nil {
		return err
	}
	requestBytes, err := c.Append(nil, request)
	if err != nil {
		return err
	}
	decoded, err := c.Decode(requestBytes)
	if err != nil {
		return err
	}
	// Offline example chooses an application outcome explicitly. No library lookup,
	// business decision, persistence or automatic unknown-user mapping occurs.
	terminal := pdu.Origin{Host: "hss.example", Realm: "example.net"}
	answer, err := c.CompleteAnswer(decoded, &s6c.SRA{AuthSessionState: 1, ExperimentalResult: avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: 5001})}, terminal)
	if err != nil {
		return err
	}
	answerBytes, err := c.Append(nil, answer)
	if err != nil {
		return err
	}
	sources := pdu.AnswerSources{Terminals: []pdu.Origin{terminal}, ErrorAgents: []pdu.Origin{{Host: "dra-a.example", Realm: "example.net"}}}
	if _, err = c.DecodeAnswer(request, answerBytes, sources); err != nil {
		return err
	}
	failure, err := c.CompleteAnswer(decoded, &pdu.ErrorAnswer{ResultCode: 3002}, sources.ErrorAgents[0])
	if err != nil {
		return err
	}
	errorBytes, err := c.Append(nil, failure)
	if err != nil {
		return err
	}
	if _, err = c.DecodeAnswer(request, errorBytes, sources); err != nil {
		return err
	}
	fmt.Printf("SRR=%d bytes SRA=%d bytes E=%d bytes; offline association verified\n", len(requestBytes), len(answerBytes), len(errorBytes))
	return nil
}
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
