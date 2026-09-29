package cmd

import (
	"testing"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/api/globalconfig"
	"github.com/evcc-io/evcc/tariff"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/text/currency"
)

func TestTariffsFromSiteReferences(t *testing.T) {
	ctrl := gomock.NewController(t)
	grid := api.NewMockTariff(ctrl)
	device := config.NewStaticDevice[api.Tariff](config.Named{Name: "multisite-grid"}, grid)
	require.NoError(t, config.Tariffs().Add(device))
	t.Cleanup(func() { require.NoError(t, config.Tariffs().Delete("multisite-grid")) })

	resolved, err := tariffsFromRefs(
		&tariff.Tariffs{Currency: currency.EUR},
		globalconfig.TariffRefs{Grid: "multisite-grid"},
	)
	require.NoError(t, err)
	assert.Same(t, grid, resolved.Grid)
	assert.Equal(t, currency.EUR, resolved.Currency)
}

func TestSiteCircuitRequiresRoot(t *testing.T) {
	ctrl := gomock.NewController(t)
	root := api.NewMockCircuit(ctrl)
	root.EXPECT().GetParent().Return(nil)
	device := config.NewStaticDevice[api.Circuit](config.Named{Name: "multisite-root"}, root)
	require.NoError(t, config.Circuits().Add(device))
	t.Cleanup(func() { require.NoError(t, config.Circuits().Delete("multisite-root")) })

	resolved, err := siteCircuit("multisite-root", false)
	require.NoError(t, err)
	assert.Same(t, root, resolved)
}

func TestScopedSiteValuesMirrorsPrimary(t *testing.T) {
	out := make(chan util.Param, 2)
	in := scopedSiteValues(out, "home", true)
	in <- util.Param{Key: "gridPower", Val: 123}

	named := <-out
	root := <-out
	assert.Equal(t, "home", named.Site)
	assert.Empty(t, root.Site)
	assert.Equal(t, named.Key, root.Key)
}
