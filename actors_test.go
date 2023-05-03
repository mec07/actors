package actors_test

import (
	"testing"
	"time"

	"github.com/mec07/actors"
)

type testActor struct{}

func (a *testActor) Receive(message interface{}) {
	return
}

type testMessage struct{}

func TestSpawnSystemWithReceiverFunc(t *testing.T) {
	expectedNumMessages := 10
	testChan := make(chan string, expectedNumMessages)
	pid := actors.SpawnSystem(actors.ReceiverFunc(func(message interface{}) {
		switch message.(type) {
		case testMessage:
			testChan <- "testMessage received"
		}
	}))

	for i := 0; i < expectedNumMessages; i++ {
		if err := pid.Send(testMessage{}); err != nil {
			t.Fatal(err)
		}
	}
	if err := pid.Send(actors.PoisonPill{}); err != nil {
		t.Fatal(err)
	}

	timer := time.NewTimer(time.Second)

	var numMessagesReceived int
	for {
		select {
		case <-testChan:
			numMessagesReceived++
			if numMessagesReceived == expectedNumMessages {
				// successful test!
				return
			}
		case <-timer.C:
			t.Fatalf("timed out waiting to get messages. Expected %d, got: %d", expectedNumMessages, numMessagesReceived)
		}
	}
}
