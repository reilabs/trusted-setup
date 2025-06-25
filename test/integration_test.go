package test

import (
	"testing"
)

func TestIntegration(t *testing.T) {
	t.Run("Test offline ceremony", TestOfflineCeremony)
}
