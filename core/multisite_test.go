package core

import (
	"testing"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/keys"
	serverdb "github.com/evcc-io/evcc/db"
	dbsettings "github.com/evcc-io/evcc/db/settings"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNamedSitesPersistSettingsIndependently(t *testing.T) {
	require.NoError(t, serverdb.NewInstance("sqlite", ":memory:"))

	home, err := NewNamedSiteFromConfig("home", nil)
	require.NoError(t, err)
	office, err := NewNamedSiteFromConfig("office", nil)
	require.NoError(t, err)

	home.SetTitle("Home")
	office.SetTitle("Office")
	require.NoError(t, dbsettings.Persist())

	homeTitle, err := dbsettings.String("site.home." + keys.Title)
	require.NoError(t, err)
	officeTitle, err := dbsettings.String("site.office." + keys.Title)
	require.NoError(t, err)
	assert.Equal(t, "Home", homeTitle)
	assert.Equal(t, "Office", officeTitle)
}

func TestSitesKeepIndependentEnergyBalances(t *testing.T) {
	ctrl := gomock.NewController(t)
	grid := func(power float64) config.Device[api.Meter] {
		meter := api.NewMockMeter(ctrl)
		meter.EXPECT().CurrentPower().Return(power, nil)
		return config.NewStaticDevice[api.Meter](config.Named{}, meter)
	}

	home := &Site{log: util.NewLogger("home"), gridMeter: grid(-2400)}
	office := &Site{log: util.NewLogger("office"), gridMeter: grid(1800)}

	homeState, err := home.updateMeters()
	require.NoError(t, err)
	officeState, err := office.updateMeters()
	require.NoError(t, err)

	assert.Equal(t, -2400.0, homeState.gridPower)
	assert.Equal(t, 1800.0, officeState.gridPower)
	assert.Equal(t, -2400.0, home.sitePower(homeState, 0, 0).power)
	assert.Equal(t, 1800.0, office.sitePower(officeState, 0, 0).power)
}
