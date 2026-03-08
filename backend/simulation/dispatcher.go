package simulation

import (
	"errors"

	"github.com/bas-x/basex/assert"
)

// Dispatcher coordinates landing requests and assignments between inbound aircraft and airbases.
type Dispatcher struct {
	constellation *Constellation
	assigner      LandingAssigner
	inbound       map[TailNumber]*InboundRecord
	inboundOrder  []TailNumber
}

// LandingAssignmentSource indicates how an assignment was produced.
type LandingAssignmentSource uint8

const (
	AssignmentSourceUnknown LandingAssignmentSource = iota
	AssignmentSourceAlgorithm
	AssignmentSourceHuman
)

// LandingAssignment captures the current base selection and provenance.
type LandingAssignment struct {
	Base   BaseID
	Source LandingAssignmentSource
}

// InboundRecord tracks a single inbound aircraft request.
type InboundRecord struct {
	Tail       TailNumber
	Assigned   bool
	Assignment LandingAssignment
}

var (
	ErrInboundNotRegistered = errors.New("dispatcher: inbound aircraft not registered")
	ErrAirbaseNotFound      = errors.New("dispatcher: airbase not found")
	ErrNoAirbases           = errors.New("dispatcher: no airbases available")
)

// LandingAssigner encapsulates assignment strategy and any internal state (for example, round robin).
type LandingAssigner interface {
	Assign(bases []Airbase, inbound []*InboundRecord, requesting TailNumber) (BaseID, error)
	Clone() LandingAssigner
}

// NewDispatcher constructs a dispatcher with the supplied constellation and landing strategy.
func NewDispatcher(constellation *Constellation, assigner LandingAssigner) *Dispatcher {
	assert.NotNil(constellation, "dispatcher constellation")
	assert.NotNil(assigner, "dispatcher assigner")
	return &Dispatcher{
		constellation: constellation,
		assigner:      assigner,
		inbound:       make(map[TailNumber]*InboundRecord),
		inboundOrder:  make([]TailNumber, 0),
	}
}

func (d *Dispatcher) AssertInvariants() {
	assert.NotNil(d, "dispatcher")
	assert.NotNil(d.constellation, "dispatcher constellation")
	assert.NotNil(d.assigner, "dispatcher assigner")
	for _, tail := range d.inboundOrder {
		record, ok := d.inbound[tail]
		assert.True(ok, "inbound record exists", tail)
		assert.NotNil(record, "inbound record")
		assert.Equal(record.Tail, tail, "inbound tail consistency")
		if record.Assigned {
			_, err := d.findAirbase(record.Assignment.Base)
			assert.Nil(err, "assigned airbase must exist")
			assert.True(record.Assignment.Source != AssignmentSourceUnknown, "assignment source")
		}
	}
}

// CloneWithConstellation replicates dispatcher state for a cloned constellation and assigner.
func (d *Dispatcher) CloneWithConstellation(constellation *Constellation) *Dispatcher {
	clone := NewDispatcher(constellation, d.assigner.Clone())
	clone.inboundOrder = append(clone.inboundOrder, d.inboundOrder...)
	for _, tail := range d.inboundOrder {
		record := d.inbound[tail]
		clonedRecord := *record
		clone.inbound[tail] = &clonedRecord
	}
	return clone
}

// RegisterInbound records or updates an inbound aircraft and produces the current assignment.
func (d *Dispatcher) RegisterInbound(tail TailNumber) (LandingAssignment, error) {
	record, exists := d.inbound[tail]
	if !exists {
		record = &InboundRecord{Tail: tail}
		d.inbound[tail] = record
		d.inboundOrder = append(d.inboundOrder, tail)
	}
	if !record.Assigned || record.Assignment.Source != AssignmentSourceHuman {
		if err := d.assign(record); err != nil {
			return record.Assignment, err
		}
	}
	return record.Assignment, nil
}

// OverrideAssignment replaces the current assignment with a human-provided choice.
func (d *Dispatcher) OverrideAssignment(tail TailNumber, base BaseID) (LandingAssignment, error) {
	record, ok := d.inbound[tail]
	if !ok {
		return LandingAssignment{}, ErrInboundNotRegistered
	}
	baseRef, err := d.findAirbase(base)
	if err != nil {
		return LandingAssignment{}, err
	}
	record.Assigned = true
	record.Assignment = LandingAssignment{Base: baseRef.ID, Source: AssignmentSourceHuman}
	return record.Assignment, nil
}

// ClearOverride reverts an inbound aircraft back to algorithmic assignment.
func (d *Dispatcher) ClearOverride(tail TailNumber) (LandingAssignment, error) {
	record, ok := d.inbound[tail]
	if !ok {
		return LandingAssignment{}, ErrInboundNotRegistered
	}
	if err := d.assign(record); err != nil {
		return LandingAssignment{}, err
	}
	return record.Assignment, nil
}

// RemoveInbound deletes an inbound record, e.g. once the aircraft commits or departs.
func (d *Dispatcher) RemoveInbound(tail TailNumber) {
	if _, exists := d.inbound[tail]; !exists {
		return
	}
	delete(d.inbound, tail)
	for i, candidate := range d.inboundOrder {
		if candidate == tail {
			last := len(d.inboundOrder) - 1
			d.inboundOrder[i] = d.inboundOrder[last]
			d.inboundOrder = d.inboundOrder[:last]
			return
		}
	}
}

// AssignmentFor returns the current assignment if any.
func (d *Dispatcher) AssignmentFor(tail TailNumber) (LandingAssignment, bool) {
	record, ok := d.inbound[tail]
	if !ok || !record.Assigned {
		return LandingAssignment{}, false
	}
	return record.Assignment, true
}

func (d *Dispatcher) assign(record *InboundRecord) error {
	bases := d.constellation.airbases
	assignment, err := d.assigner.Assign(bases, d.InboundRecords(), record.Tail)
	if err != nil {
		record.Assigned = false
		record.Assignment = LandingAssignment{}
		return err
	}
	record.Assigned = true
	record.Assignment = LandingAssignment{Base: assignment, Source: AssignmentSourceAlgorithm}
	return nil
}

func (d *Dispatcher) findAirbase(id BaseID) (*Airbase, error) {
	bases := d.constellation.airbases
	for i := range bases {
		if bases[i].ID == id {
			return &bases[i], nil
		}
	}
	return nil, ErrAirbaseNotFound
}

// InboundRecords returns the deterministic slice of inbound records.
func (d *Dispatcher) InboundRecords() []*InboundRecord {
	result := make([]*InboundRecord, 0, len(d.inboundOrder))
	for _, tail := range d.inboundOrder {
		if record, ok := d.inbound[tail]; ok {
			result = append(result, record)
		}
	}
	return result
}

// RoundRobinAssigner cycles through airbases sequentially.
type RoundRobinAssigner struct {
	next int
}

// Assign implements LandingAssigner.
func (r *RoundRobinAssigner) Assign(bases []Airbase, _ []*InboundRecord, _ TailNumber) (BaseID, error) {
	if len(bases) == 0 {
		return BaseID{}, ErrNoAirbases
	}
	index := r.next % len(bases)
	r.next = (r.next + 1) % len(bases)
	return bases[index].ID, nil
}

// Clone returns a deep copy of the assigner.
func (r *RoundRobinAssigner) Clone() LandingAssigner {
	if r == nil {
		return &RoundRobinAssigner{}
	}
	return &RoundRobinAssigner{next: r.next}
}
