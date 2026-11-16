// Package diameter provides bounded Diameter endpoints for public Go callers.
//
// Nodes own peers, transport candidates, completion ports and resource budgets.
// Explicit command policies bind application profiles, terminal destinations and
// allowed answer sources independently of an adjacent DRA. Submit is asynchronous;
// CompletionPort.Receive consumes the retained result and releases its charge.
// Inbound.Respond validates an answer and restores the original arrival envelope.
// Optional, finite in-memory duplicate caching supports multiple arrivals; it
// provides no durable business or financial deduplication.
//
// Dial/Maintain and Listener/Serve own transport admission and cancellation.
// TCP/TLS adapters live in this module; Linux SCTP is in the separate adapters/sctp
// module. Shutdown drains and uses DPR/DPA, Close terminates I/O, and Wait joins
// cleanup. Completion ports must also be consumed or closed to release resources.
//
// Generated, owned codecs provide selected Base, Sh, S6c, SGd/Gdd and Ro/DCCA
// profiles, plus explicit sms-fw compatibility bindings. They are bounded subsets,
// not a declaration of complete protocol or target-OCS conformance. Sh-Data helpers
// are explicit; codecs do not implicitly interpret business XML or SMS payloads.
//
// Legacy SCAP1 App0/271 needs SCAP1Accounting policies and strict local Vendor193 /
// Acct19302 capabilities. PeerConfig.SCAP1Advertisement optionally selects the
// reviewed old dual advertisement; standard Base validation remains the default.
// SCAP1 remains SingleAttempt. Selected CCR bindings have separate, explicit Ro
// recovery APIs; these do not authorize SCAP1 recovery or automatic financial retry.
//
// All production packages and dependencies are pure Go, targeting Go 1.26+.
// The library has no persistence, subscriber lookup, charging ledger or application
// restart orchestration. Current implementation boundaries are in spec/progress.json.
package diameter

//go:generate go run ./cmd/diameter-schema -in spec/base.json -out base/zz_generated_messages.go
//go:generate go run ./cmd/diameter-schema -in spec/s6c.json -out app/s6c/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/sgd.json -out app/sgd/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/sh.json -out app/sh/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/ro.json -out app/ro/zz_generated_messages.go
//go:generate go run ./cmd/diameter-schema -in spec/roabort.json -out app/ro/abort/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/smsfwro.json -out compat/smsfwro/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/legacy420.json -out compat/smsfwro/legacy420/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/scap1.json -out compat/scap1/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/combined.json -out compat/smsfwro/combined/zz_generated_messages.go

//go:generate go run ./cmd/diameter-schema -in spec/dcca.json -out app/ro/dcca/zz_generated_messages.go
