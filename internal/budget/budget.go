// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package budget provides a single-charge reservation mechanism. Its owner must
// serialize all access; the mechanism never performs I/O or invokes callbacks.
package budget

import "errors"

var ErrLimit = errors.New("resource budget exhausted")

type Account struct {
	LimitBytes, UsedBytes uint64
	LimitSlots, UsedSlots uint64
}
type Reservation struct {
	Capacity, Held uint64
	Owner          string
	slots          uint64
	accounts       [4]*Account
	count          int
	released       bool
}

func Reserve(bytes, slots uint64, owner string, accounts ...*Account) (*Reservation, error) {
	if len(accounts) == 0 || len(accounts) > 4 {
		return nil, ErrLimit
	}
	for i, a := range accounts {
		if a == nil || a.UsedBytes > a.LimitBytes || bytes > a.LimitBytes-a.UsedBytes || a.UsedSlots > a.LimitSlots || slots > a.LimitSlots-a.UsedSlots {
			return nil, ErrLimit
		}
		for _, b := range accounts[:i] {
			if a == b {
				return nil, ErrLimit
			}
		}
	}
	r := &Reservation{Capacity: bytes, Owner: owner, slots: slots, count: len(accounts)}
	copy(r.accounts[:], accounts)
	for _, a := range accounts {
		a.UsedBytes += bytes
		a.UsedSlots += slots
	}
	return r, nil
}
func (r *Reservation) SetHeld(n uint64) error {
	if r.released || n > r.Capacity {
		return ErrLimit
	}
	r.Held = n
	return nil
}
func (r *Reservation) Transfer(owner string) {
	if !r.released {
		r.Owner = owner
	}
}
func (r *Reservation) Shrink(n uint64) error {
	if r.released || n > r.Capacity || n < r.Held {
		return ErrLimit
	}
	for _, a := range r.accounts[:r.count] {
		a.UsedBytes -= r.Capacity - n
	}
	r.Capacity = n
	return nil
}

// Split moves unused capacity without another charge. Slots stay with the parent.
func (r *Reservation) Split(n uint64, owner string) (*Reservation, error) {
	if r.released || n > r.Capacity-r.Held {
		return nil, ErrLimit
	}
	child := &Reservation{Capacity: n, Owner: owner, accounts: r.accounts, count: r.count}
	r.Capacity -= n
	return child, nil
}
func (r *Reservation) Detach(a *Account) {
	if r.released {
		return
	}
	for i, v := range r.accounts[:r.count] {
		if v == a {
			a.UsedBytes -= r.Capacity
			a.UsedSlots -= r.slots
			r.count--
			r.accounts[i] = r.accounts[r.count]
			r.accounts[r.count] = nil
			return
		}
	}
}
func (r *Reservation) Release() {
	if r == nil || r.released {
		return
	}
	r.released = true
	for _, a := range r.accounts[:r.count] {
		a.UsedBytes -= r.Capacity
		a.UsedSlots -= r.slots
	}
	r.Capacity = 0
	r.Held = 0
	r.Owner = "released"
}
