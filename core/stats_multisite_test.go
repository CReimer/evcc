package core

import (
	"testing"
	"time"

	"github.com/evcc-io/evcc/core/session"
	serverdb "github.com/evcc-io/evcc/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsFilterNamedSiteSessions(t *testing.T) {
	require.NoError(t, serverdb.NewInstance("sqlite", ":memory:"))
	solar := 0.5
	require.NoError(t, serverdb.Instance.Create(&[]session.Session{
		{Finished: time.Now(), Site: "home_1", Loadpoint: "garage", ChargedEnergy: 10, SolarPercentage: &solar},
		{Finished: time.Now(), Site: "homeA1", Loadpoint: "parking", ChargedEnergy: 30, SolarPercentage: &solar},
	}).Error)

	assert.Equal(t, 10.0, NewStats("home_1").calculate(time.Time{})["chargedKWh"])
	assert.Equal(t, 30.0, NewStats("homeA1").calculate(time.Time{})["chargedKWh"])
	assert.Equal(t, 40.0, NewStats("").calculate(time.Time{})["chargedKWh"])
}
