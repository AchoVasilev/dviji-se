package config

import (
	"os"
	"sync"
	"testing"
)

// reset clears the once-guarded singleton so a test can load a fresh config.
func reset() {
	c = nil
	once = sync.Once{}
}

// Workouts and nutrition have no pages yet, so a deployment that says nothing
// about them must not advertise them.
func TestFeatureFlags_Defaults(t *testing.T) {
	for _, key := range []string{"ENABLED_WORKOUTS", "ENABLED_NUTRITION"} {
		os.Unsetenv(key)
	}

	t.Setenv("JWT_KEY", "k")
	t.Setenv("JWT_REFRESH_KEY", "k")
	t.Setenv("XSRF", "k")

	reset()
	defer reset()

	if WorkoutsEnabled() || NutritionEnabled() {
		t.Error("sections without pages should default to hidden")
	}

}

func TestFeatureFlags_ToggleIndependently(t *testing.T) {
	t.Setenv("JWT_KEY", "k")
	t.Setenv("JWT_REFRESH_KEY", "k")
	t.Setenv("XSRF", "k")
	t.Setenv("ENABLED_WORKOUTS", "true")
	t.Setenv("ENABLED_NUTRITION", "false")

	reset()
	defer reset()

	if !WorkoutsEnabled() {
		t.Error("ENABLED_WORKOUTS=true should enable workouts")
	}

	if NutritionEnabled() {
		t.Error("ENABLED_NUTRITION=false should keep nutrition hidden")
	}

}
