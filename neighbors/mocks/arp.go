// Code generated by Clue Mock Generator v1.1.1, DO NOT EDIT.
//
// Command:
// $ cmg gen douglasthrift.net/presence/neighbors

package mockneighbors

import (
	"context"
	"testing"

	"goa.design/clue/mock"

	"douglasthrift.net/presence/neighbors"
)

type (
	ARP struct {
		m *mock.Mock
		t *testing.T
	}

	ARPPresentFunc func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error
	ARPCountFunc   func(count uint)
)

func NewARP(t *testing.T) *ARP {
	var (
		m               = &ARP{mock.New(), t}
		_ neighbors.ARP = m
	)
	return m
}

func (m *ARP) AddPresent(f ARPPresentFunc) {
	m.m.Add("Present", f)
}

func (m *ARP) SetPresent(f ARPPresentFunc) {
	m.m.Set("Present", f)
}

func (m *ARP) Present(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
	if f := m.m.Next("Present"); f != nil {
		return f.(ARPPresentFunc)(ctx, ifs, state, addrStates)
	}
	m.t.Helper()
	m.t.Error("unexpected Present call")
	return nil
}

func (m *ARP) AddCount(f ARPCountFunc) {
	m.m.Add("Count", f)
}

func (m *ARP) SetCount(f ARPCountFunc) {
	m.m.Set("Count", f)
}

func (m *ARP) Count(count uint) {
	if f := m.m.Next("Count"); f != nil {
		f.(ARPCountFunc)(count)
		return
	}
	m.t.Helper()
	m.t.Error("unexpected Count call")
}

func (m *ARP) HasMore() bool {
	return m.m.HasMore()
}
