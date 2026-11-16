package shdata_test

import (
	"fmt"

	"github.com/skill215/diameterkit/app/sh/shdata"
)

func ExampleParse() {
	limits := shdata.DefaultLimits()
	raw, err := shdata.MarshalSMSUpdate(shdata.SMSRegistration{Number: "12025550999"}, limits)
	if err != nil {
		panic(err)
	}
	// raw can be assigned explicitly to sh.PUR.UserData. Neither codec.Decode
	// nor Node/Peer invokes XML parsing or chooses the application result.
	doc, err := shdata.Parse(raw, limits)
	if err != nil {
		panic(err)
	}
	report, err := doc.Check(shdata.SMSUpdate)
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Data().SMS.Value.Number, report.Complete)
	// Forward doc.Raw() to retain exact bytes; rebuilding may change lexical form.
	// Output: 12025550999 true
}
