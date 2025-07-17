package randomness

// MockProvider implements BeaconProvider and returns a fixed 32-byte beacon.
type MockProvider struct {
	Beacon []byte
}

// GetBeacon returns the 32 bytes of randomness acquired at the module initialization.
func (m *MockProvider) GetBeacon() []byte {
	return m.Beacon
}
