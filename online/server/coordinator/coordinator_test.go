package coordinator_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/contribution"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager/unique_fifo"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
)

type mockPhase2 struct {
	phase2 bytes.Buffer
	count  int
}

func (m *mockPhase2) GetCount() int {
	return m.count
}

func (m *mockPhase2) NewVerifiable() contribution.Verifiable {
	return &mockPhase2{}
}

func (m *mockPhase2) GetContribution() interface{} {
	return &m.phase2
}

func (m *mockPhase2) AddContribution(next contribution.Verifiable) error {
	contribBuf := next.(*mockPhase2).phase2
	// Simulate an error mpcsetup.Verify could return if it dislikes the contribution
	if contribBuf.Len() != 0x37 {
		return errors.New("malformed contribution")
	}
	_, err := m.phase2.Write(contribBuf.Bytes())
	m.count++
	return err
}

func (m *mockPhase2) ExtractKeys() (groth16.ProvingKey, groth16.VerifyingKey) {
	// Not necessary for this test
	panic("not implemented")
}

func (m *mockPhase2) ReadFrom(reader io.Reader) (int64, error) {
	return m.phase2.ReadFrom(reader)
}

func (m *mockPhase2) WriteTo(writer io.Writer) (int64, error) {
	fakeContrib := bytes.NewBuffer(bytes.Repeat([]byte{0x21}, 0x37))
	return fakeContrib.WriteTo(writer)
}

type mockContributorsManager struct {
	contributorsCount int
}

func (m *mockContributorsManager) AddContributor(notify contributors_manager.OnPositionUpdate) contributors_manager.PositionUpdateNotifier {
	m.contributorsCount++
	return func() {
		notify(m.contributorsCount - 1)
	}
}

func (m *mockContributorsManager) RemoveCurrentContributor() error {
	if m.contributorsCount == 0 {
		return unique_fifo.ErrEmpty
	}
	// Noop, let's pretend the contributor was removed
	return nil
}

func TestAddContributor(t *testing.T) {
	coord := coordinator.New(&mockPhase2{}, &mockContributorsManager{})

	onPositionUpdate0 := coord.AddContributor(
		func(position int) {
			assert.Equal(t, 0, position)
		},
	)
	assert.NotNil(t, onPositionUpdate0)
	onPositionUpdate0()

	onPositionUpdate1 := coord.AddContributor(
		func(position int) {
			assert.Equal(t, 1, position)
		},
	)
	assert.NotNil(t, onPositionUpdate1)
	onPositionUpdate1()
}

func TestWriteLastContribution(t *testing.T) {
	coord := coordinator.New(&mockPhase2{}, &mockContributorsManager{})

	var clientBuf bytes.Buffer

	_ = coord.AddContributor(func(int) {})
	n, err := coord.WriteLastContribution(&clientBuf)
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0x37), n) // known fake contribution size
	assert.Equal(t, n, int64(clientBuf.Len()), n)
}

func TestReadNextContribution(t *testing.T) {
	coord := coordinator.New(&mockPhase2{}, &mockContributorsManager{})
	assert.Equal(t, 0, coord.GetContributionsCount())

	_ = coord.AddContributor(func(int) {})

	goodContrib := bytes.NewBuffer(bytes.Repeat([]byte{0x21}, 0x37))
	_, err := coord.ReadNextContribution(goodContrib)
	assert.NoError(t, err)
	assert.Equal(t, 1, coord.GetContributionsCount())

	// Shortcut here - during the real ceremony the contributor would be
	// removed from the queue, and we'd had to test bad contribution with
	// another contributor. This test however does not implement removal,
	// so we can accept more contributions from one contributor and test
	// for other conditions.
	badContrib := bytes.NewBuffer([]byte{0x1, 0x01})
	_, err = coord.ReadNextContribution(badContrib)
	assert.Error(t, err)
	assert.Equal(t, 1, coord.GetContributionsCount())
}
