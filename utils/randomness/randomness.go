// Package randomness provides "publicly verifiable, unbiased and unpredictable random values" as guaranteed
// by the drand organization (https://github.com/drand/drand).
package randomness

// BeaconProvider defines an interface to get 32 bytes of randomness (beacon).
type BeaconProvider interface {
	GetBeacon() []byte
}
