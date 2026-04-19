package version

import "testing"

func TestGetReturnsSomething(t *testing.T) {
	info := Get()
	if info.Version == "" || info.Commit == "" || info.BuildDate == "" {
		t.Errorf("all fields should be populated, got %+v", info)
	}
}
