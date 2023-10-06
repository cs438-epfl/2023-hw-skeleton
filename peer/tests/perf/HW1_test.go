//go:build performance
// +build performance

package perf

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	z "go.dedis.ch/cs438/internal/testing"
	"go.dedis.ch/cs438/registry/standard"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

func Test_HW1_BenchmarkSpamNode(t *testing.T) {
	// run the benchmark
	res := testing.Benchmark(BenchmarkSpamNode)

	// assess allocation against thresholds, the performance thresholds is the allocation on GitHub
	assessAllocs(t, res, []allocThresholds{
		{"allocs great", 90, 15_000},
		{"allocs ok", 140, 22_000},
		{"allocs passable", 180, 30_000},
	})

	// assess execution speed against thresholds, the performance thresholds is the execution speed on GitHub
	assessSpeed(t, res, []speedThresholds{
		{"speed great", 20 * time.Millisecond},
		{"speed ok", 100 * time.Millisecond},
		{"speed passable", 600 * time.Millisecond},
	})
}

// Flood a node with n messages and wait on rumors to be processed. Rumors are
// sent with expected sequences, so they should be processed. From the root
// folder, you can run the benchmark as follow:
//
//	GLOG=no go test --bench BenchmarkSpamNode -benchtime 1000x -benchmem \
//		-count 5 ./peer/tests/extra
func BenchmarkSpamNode(b *testing.B) {
	// Disable outputs to not penalize implementations that make use of it
	oldStdout := os.Stdout
	os.Stdout = nil

	defer func() {
		os.Stdout = oldStdout
	}()

	n := b.N
	transp := channelFac()

	fake := z.NewFakeMessage(b)

	notifications := make(chan struct{}, n)

	handler := func(types.Message, transport.Packet) error {
		notifications <- struct{}{}
		return nil
	}

	sender, err := transp.CreateSocket("127.0.0.1:0")
	require.NoError(b, err)

	defer sender.Close()

	receiver := z.NewTestNode(b, peerFac, transp, "127.0.0.1:0", z.WithMessage(fake, handler), z.WithContinueMongering(0))
	defer receiver.Stop()

	src := sender.GetAddress()
	dst := receiver.GetAddr()

	currentRumorSeq := 0
	acketPackedID := make([]byte, 12)

	_, err = rand.Read(acketPackedID)
	require.NoError(b, err)

	for i := 0; i < n; i++ {
		// flip a coin to send either a rumor or an ack message
		coin := rand.Float64() > 0.5

		if coin {
			currentRumorSeq++

			err = sendRumor(fake, uint(currentRumorSeq), src, dst, sender)
			require.NoError(b, err)
		} else {
			err = sendAck(string(acketPackedID), src, dst, sender)
			require.NoError(b, err)
		}
	}

	// > wait on all the rumors to be processed

	for i := 0; i < currentRumorSeq; i++ {
		select {
		case <-notifications:
		case <-time.After(time.Second):
			b.Error("notification not received in time")
		}
	}
}

func sendRumor(msg types.Message, seq uint, src, dst string, sender transport.Socket) error {

	embedded, err := defaultRegistry.MarshalMessage(msg)
	if err != nil {
		return xerrors.Errorf("failed to marshal msg: %v", err)
	}

	rumor := types.Rumor{
		Origin:   src,
		Sequence: seq,
		Msg:      &embedded,
	}

	rumorsMessage := types.RumorsMessage{
		Rumors: []types.Rumor{rumor},
	}

	transpMsg, err := defaultRegistry.MarshalMessage(rumorsMessage)
	if err != nil {
		return xerrors.Errorf("failed to marshal transp msg: %v", err)
	}

	header := transport.NewHeader(src, src, dst, 1)

	pkt := transport.Packet{
		Header: &header,
		Msg:    &transpMsg,
	}

	err = sender.Send(dst, pkt, 0)
	if err != nil {
		return xerrors.Errorf("failed to send message: %v", err)
	}

	return nil
}

func sendAck(ackedPacketID string, src, dst string, sender transport.Socket) error {
	ack := types.AckMessage{
		AckedPacketID: ackedPacketID,
	}

	registry := standard.NewRegistry()

	msg, err := registry.MarshalMessage(ack)
	if err != nil {
		return xerrors.Errorf("failed to marshal message: %v", err)
	}

	header := transport.NewHeader(src, src, dst, 1)

	pkt := transport.Packet{
		Header: &header,
		Msg:    &msg,
	}

	err = sender.Send(dst, pkt, 0)
	if err != nil {
		return xerrors.Errorf("failed to send message: %v", err)
	}

	return nil
}
