package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNeed_Degrade(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		start  int
		amount int
		want   int
	}{
		{name: "increase severity", start: 30, amount: 5, want: 35},
		{name: "cap at one hundred", start: 98, amount: 5, want: 100},
		{name: "zero is no op", start: 42, amount: 0, want: 42},
		{name: "negative is no op", start: 42, amount: -5, want: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			need := Need{Type: NeedFuel, Severity: tt.start, RequiredCapability: NeedFuel}
			need.Degrade(tt.amount)

			require.Equal(t, tt.want, need.Severity)
		})
	}
}

func TestNeed_Restore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		start  int
		amount int
		want   int
	}{
		{name: "decrease severity", start: 30, amount: 5, want: 25},
		{name: "cap at zero", start: 3, amount: 5, want: 0},
		{name: "zero is no op", start: 42, amount: 0, want: 42},
		{name: "negative is no op", start: 42, amount: -5, want: 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			need := Need{Type: NeedFuel, Severity: tt.start, RequiredCapability: NeedFuel}
			need.Restore(tt.amount)

			require.Equal(t, tt.want, need.Severity)
		})
	}
}

func TestNeedsThresholdReached(t *testing.T) {
	t.Parallel()
	threshold := testLifecycleModel().ReturnThreshold

	require.False(t, NeedsThresholdReached(nil, threshold))
	require.False(t, NeedsThresholdReached([]Need{}, threshold))
	require.False(t, NeedsThresholdReached([]Need{
		{Type: NeedFuel, Severity: 79, RequiredCapability: NeedFuel},
		{Type: NeedMunitions, Severity: 12, RequiredCapability: NeedMunitions},
	}, threshold))
	require.True(t, NeedsThresholdReached([]Need{
		{Type: NeedFuel, Severity: 80, RequiredCapability: NeedFuel},
		{Type: NeedMunitions, Severity: 12, RequiredCapability: NeedMunitions},
	}, threshold))
}

func TestNeedProgressionDeterministic(t *testing.T) {
	t.Parallel()

	lifecycle := testLifecycleModel()
	a := NewAircraft(TailNumber{1, 2, 3}, &OutboundState{}, []Need{{Type: NeedFuel, Severity: 20, RequiredCapability: NeedFuel}})
	b := NewAircraft(TailNumber{1, 2, 3}, &OutboundState{}, []Need{{Type: NeedFuel, Severity: 20, RequiredCapability: NeedFuel}})

	a.ApplyNeedPhase(30*time.Second, NeedPhaseOutbound, lifecycle, nil)
	b.ApplyNeedPhase(30*time.Second, NeedPhaseOutbound, lifecycle, nil)

	require.Equal(t, a.Needs, b.Needs)
	require.Equal(t, a.NeedRemainders, b.NeedRemainders)
}
