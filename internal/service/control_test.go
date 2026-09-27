package service

import (
	"context"
	"math"
	"testing"
)

func TestLaunchRejectsNonFiniteTargetPressure(t *testing.T) {
	cases := map[string]float64{
		"NaN":         math.NaN(),
		"positiveInf": math.Inf(1),
		"negativeInf": math.Inf(-1),
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			runtime := NewRuntime()
			_, err := runtime.Launch(context.Background(), LaunchRequest{
				StationID:      "station-a",
				RouteID:        "main-station-a-station-c",
				TargetPressure: target,
			})
			if err == nil {
				t.Fatalf("expected launch with %s target pressure to be rejected", name)
			}
			if got := len(runtime.Operations()); got != 0 {
				t.Fatalf("expected no runs after rejection, got %d", got)
			}
			snapshot := runtime.Snapshot()
			for _, routeState := range snapshot.Routes {
				if routeState.Reserved {
					t.Fatalf("route %s reserved after rejection", routeState.ID)
				}
			}
			for _, pressureState := range snapshot.Pressures {
				if math.IsNaN(pressureState.Value) || math.IsInf(pressureState.Value, 0) {
					t.Fatalf("compressor at %s holds non-finite pressure %v", pressureState.StationID, pressureState.Value)
				}
				if pressureState.RunID != "" {
					t.Fatalf("compressor at %s bound to run %s after rejection", pressureState.StationID, pressureState.RunID)
				}
			}
			if err := snapshot.Validate(); err != nil {
				t.Fatalf("snapshot invalid after rejection: %v", err)
			}
		})
	}
}

func TestLaunchAcceptsFiniteTargetPressure(t *testing.T) {
	runtime := NewRuntime()
	operation, err := runtime.Launch(context.Background(), LaunchRequest{
		StationID:      "station-a",
		RouteID:        "main-station-a-station-c",
		TargetPressure: 10.5,
	})
	if err != nil {
		t.Fatalf("expected finite target pressure to launch, got %v", err)
	}
	if operation.Run.ID == "" {
		t.Fatal("expected a run to be created")
	}
	if operation.Pressure.Value != 10.5 || operation.Pressure.Target != 10.5 {
		t.Fatalf("expected compressor holding 10.5, got value %.2f target %.2f", operation.Pressure.Value, operation.Pressure.Target)
	}
}
