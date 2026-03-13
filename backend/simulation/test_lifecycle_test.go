package simulation

import "time"

func testLifecycleModel() LifecycleModel {
	return LifecycleModel{
		Durations: PhaseDurations{
			Outbound:        5 * time.Second,
			Engaged:         5 * time.Second,
			InboundDecision: 3 * time.Second,
			CommitApproach:  4 * time.Second,
			Servicing:       6 * time.Second,
			Ready:           2 * time.Second,
		},
		ReturnThreshold: 80,
		NeedRates: map[NeedType]NeedRateModel{
			NeedFuel:      {OutboundMilliPerHour: 18000000, EngagedMilliPerHour: 18000000, ServicingMilliPerHour: 28800000, VariancePermille: 0},
			NeedMunitions: {OutboundMilliPerHour: 18000000, EngagedMilliPerHour: 18000000, ServicingMilliPerHour: 18000000, VariancePermille: 0},
		},
	}
}
