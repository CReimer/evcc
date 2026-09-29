package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPersistSubscribers(t *testing.T) {
	persistSubscribers.Lock()
	persistSubscribers.callbacks = nil
	persistSubscribers.Unlock()
	t.Cleanup(func() {
		persistSubscribers.Lock()
		persistSubscribers.callbacks = nil
		persistSubscribers.Unlock()
	})

	var calls int
	SubscribePersist(func(time.Time) { calls++ })
	SubscribePersist(func(time.Time) { calls++ })
	notifyPersist(time.Now())

	assert.Equal(t, 2, calls)
}
