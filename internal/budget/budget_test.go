package budget

import "testing"

func TestAtomicDimensionsAndTransfer(t *testing.T) {
	node := &Account{LimitBytes: 100, LimitSlots: 2}
	class := &Account{LimitBytes: 80, LimitSlots: 2}
	peer := &Account{LimitBytes: 70, LimitSlots: 1}
	port := &Account{LimitBytes: 80, LimitSlots: 2}
	r, err := Reserve(60, 1, "exchange", node, class, peer, port)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Reserve(10, 1, "other", node, class, peer, port); err == nil {
		t.Fatal("peer window bypass")
	}
	if node.UsedBytes != 60 || port.UsedSlots != 1 {
		t.Fatal("partial reservation")
	}
	if err = r.SetHeld(20); err != nil {
		t.Fatal(err)
	}
	if _, err = r.Split(41, "bad"); err == nil {
		t.Fatal("split borrowed")
	}
	child, err := r.Split(30, "answer")
	if err != nil {
		t.Fatal(err)
	}
	if node.UsedBytes != 60 {
		t.Fatal("double charge")
	}
	child.Detach(peer)
	child.Transfer("completion")
	if peer.UsedBytes != 30 || node.UsedBytes != 60 {
		t.Fatal("detach")
	}
	r.Release()
	r.Release()
	if node.UsedBytes != 30 || peer.UsedSlots != 0 || node.UsedSlots != 0 {
		t.Fatal("release")
	}
	child.Release()
	if node.UsedBytes != 0 || port.UsedBytes != 0 {
		t.Fatal("leak")
	}
}
func TestBudgetOverflowAndShrink(t *testing.T) {
	a := &Account{LimitBytes: ^uint64(0), LimitSlots: 1}
	r, err := Reserve(^uint64(0)-1, 1, "x", a)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Reserve(2, 0, "x", a); err == nil {
		t.Fatal("overflow")
	}
	if err = r.SetHeld(5); err != nil {
		t.Fatal(err)
	}
	if err = r.Shrink(4); err == nil {
		t.Fatal("shrunk held")
	}
	if err = r.Shrink(7); err != nil {
		t.Fatal(err)
	}
	r.Release()
	if a.UsedBytes != 0 {
		t.Fatal("leak")
	}
}
