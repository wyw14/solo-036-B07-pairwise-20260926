package service

import (
	"context"
	"math"
	"strings"
	"testing"
)

func launchRequest(target float64) LaunchRequest {
	return LaunchRequest{StationID: "station-a", RouteID: "main-station-a-station-c", TargetPressure: target}
}

func TestLaunchRejectsNonFinitePressure(t *testing.T) {
	cases := map[string]float64{
		"NaN":        math.NaN(),
		"+Inf":       math.Inf(1),
		"-Inf":       math.Inf(-1),
		"zero":       0,
		"negative":   -3.5,
		"over limit": 12.6,
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			runtime := NewRuntime()
			if _, err := runtime.Launch(context.Background(), launchRequest(target)); err == nil {
				t.Fatalf("launch with target %v succeeded, want rejection", target)
			}
			if got := len(runtime.Operations()); got != 0 {
				t.Fatalf("rejected launch created %d runs, want 0", got)
			}
			for _, lease := range runtime.Snapshot().Routes {
				if lease.Reserved {
					t.Fatalf("rejected launch left route %s reserved", lease.ID)
				}
			}
			if current := runtime.Pressure.Current(); current != 0 || math.IsNaN(current) || math.IsInf(current, 0) {
				t.Fatalf("compressor holds invalid pressure %v after rejection", current)
			}
		})
	}
}

func TestLaunchRejectsNonFinitePressureMessage(t *testing.T) {
	runtime := NewRuntime()
	err := runtime.ValidateLaunchRequest(launchRequest(math.NaN()))
	if err == nil || !strings.Contains(err.Error(), "finite") {
		t.Fatalf("validation error = %v, want finite-pressure rejection", err)
	}
}

func TestLaunchAcceptsFinitePressure(t *testing.T) {
	runtime := NewRuntime()
	operation, err := runtime.Launch(context.Background(), launchRequest(10.5))
	if err != nil {
		t.Fatalf("launch with finite target rejected: %v", err)
	}
	if operation.Run.ID == "" {
		t.Fatal("launch returned no run")
	}
	if current := runtime.Pressure.Current(); current != 10.5 {
		t.Fatalf("compressor holds %v, want 10.5", current)
	}
}
