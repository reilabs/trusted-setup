package coordinator_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/stretchr/testify/assert"

	"github.com/reilabs/trusted-setup/online/phase2"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager"
	"github.com/reilabs/trusted-setup/online/server/contributors_manager/unique_fifo"
	"github.com/reilabs/trusted-setup/online/server/coordinator"
)

type mockPhase2 struct {
	phase2 bytes.Buffer
}

func (m *mockPhase2) NewVerifiable() phase2.Verifiable {
	return &mockPhase2{}
}

func (m *mockPhase2) GetContribution() interface{} {
	return &m.phase2
}

func (m *mockPhase2) AddContribution(next phase2.Verifiable) error {
	contribBuf := next.GetContribution().(*bytes.Buffer)
	// Simulate an error mpcsetup.Verify could return if it dislikes the contribution
	if contribBuf.Len() != 0x37 {
		return errors.New("malformed contribution")
	}
	_, err := m.phase2.Write(contribBuf.Bytes())
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
	contributors []string
}

func (m *mockContributorsManager) AddContributor(contributorIp string) contributors_manager.PositionUpdateNotifier {
	m.contributors = append(m.contributors, contributorIp)
	return func(onPositionUpdate contributors_manager.OnPositionUpdate) {
		onPositionUpdate(len(m.contributors) - 1)
	}
}

func (m *mockContributorsManager) RemoveCurrentContributor() error {
	if len(m.contributors) == 0 {
		return unique_fifo.ErrEmpty
	}
	// Noop, let's pretend the contributor was removed
	return nil
}

func (m *mockContributorsManager) EnsureNextContributor(candidateIp string) error {
	if len(m.contributors) == 0 {
		return unique_fifo.ErrEmpty
	}
	if m.contributors[0] != candidateIp {
		return contributors_manager.ErrContributorOutOfOrder
	}
	return nil
}

func TestAddContributor(t *testing.T) {
	coord := coordinator.New(&mockPhase2{}, &mockContributorsManager{})

	onPositionUpdate0 := coord.AddContributor("127.0.0.1")
	assert.NotNil(t, onPositionUpdate0)
	onPositionUpdate0(
		func(position int) {
			assert.Equal(t, 0, position)
		},
	)

	onPositionUpdate1 := coord.AddContributor("127.0.0.2")
	assert.NotNil(t, onPositionUpdate1)
	onPositionUpdate1(
		func(position int) {
			assert.Equal(t, 1, position)
		},
	)
}

func TestUploadContributionToClient(t *testing.T) {
	coord := coordinator.New(&mockPhase2{}, &mockContributorsManager{})

	var clientBuf bytes.Buffer

	_ = coord.AddContributor("127.0.0.1")
	n, err := coord.WriteLastContribution("127.0.0.2", &clientBuf)
	assert.ErrorIs(t, err, contributors_manager.ErrContributorOutOfOrder)
	assert.Equal(t, int64(0), n)

	n, err = coord.WriteLastContribution("127.0.0.1", &clientBuf)
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0x37), n) // known fake contribution size
	assert.Equal(t, n, int64(clientBuf.Len()), n)
}

func TestCoordinator_DownloadContributionFromClient(t *testing.T) {
	coord := coordinator.New(&mockPhase2{}, &mockContributorsManager{})

	fakeContrib := bytes.NewBuffer(bytes.Repeat([]byte{0x21}, 0x37))

	_ = coord.AddContributor("127.0.0.1")
	n, err := coord.ReadNextContribution("127.0.0.2", fakeContrib)
	assert.ErrorIs(t, err, contributors_manager.ErrContributorOutOfOrder)
	assert.Equal(t, int64(0), n)

	n, err = coord.ReadNextContribution("127.0.0.1", fakeContrib)
	assert.NoError(t, err)
	assert.Equal(t, n, int64(0x37), n) // known fake contribution size
	fakeContrib = bytes.NewBuffer(bytes.Repeat([]byte{0x21}, 0x37))
}

func TestVerifyContribution(t *testing.T) {
	coord := coordinator.New(&mockPhase2{}, &mockContributorsManager{})

	_ = coord.AddContributor("127.0.0.1")

	goodContrib := bytes.NewBuffer(bytes.Repeat([]byte{0x21}, 0x37))
	_, err := coord.ReadNextContribution("127.0.0.1", goodContrib)
	assert.NoError(t, err)

	err = coord.VerifyNextContribution()
	assert.NoError(t, err)

	// Shortcut here - during the real ceremony the contributor would be
	// removed from the queue, and we'd had to test bad contribution with
	// another contributor. This test however does not implement removal,
	// so we can accept more contributions from one contributor and test
	// for other conditions.
	badContrib := bytes.NewBuffer([]byte{0x1, 0x01})
	_, err = coord.ReadNextContribution("127.0.0.1", badContrib)
	assert.NoError(t, err)

	err = coord.VerifyNextContribution()
	assert.Error(t, err)
}
