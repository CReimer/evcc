package cmd

import (
	"strings"
	"testing"

	"github.com/evcc-io/evcc/api/globalconfig"
	"github.com/evcc-io/evcc/core/keys"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
	spfViper "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopedSiteValuesKeepsVehiclesGlobal(t *testing.T) {
	out := make(chan util.Param)
	in := scopedSiteValues(out, "home", true)
	in <- util.Param{Key: keys.Vehicles, Val: "vehicles"}
	param := <-out
	assert.Empty(t, param.Site)
	assert.Equal(t, "vehicles", param.Val)

	secondary := scopedSiteValues(out, "office", false)
	secondary <- util.Param{Key: keys.Vehicles, Val: "duplicate"}
	secondary <- util.Param{Key: keys.PvPower, Val: 1200}
	param = <-out
	assert.Equal(t, "office", param.Site)
	assert.Equal(t, keys.PvPower, param.Key)
}

func TestSitesYaml(t *testing.T) {
	spfViper.Reset()
	t.Cleanup(spfViper.Reset)
	spfViper.SetConfigType("yaml")
	require.NoError(t, spfViper.ReadConfig(strings.NewReader(`
sites:
  - name: home
    title: Home
    circuit: home-root
    tariffs:
      grid: home-grid-tariff
    hems:
      type: custom
      dim: true
    meters:
      grid: home-grid
  - name: office
    title: Office
    meters:
      grid: office-grid
loadpoints:
  - name: garage
    site: home
  - name: parking
    site: office
`)))

	var conf globalconfig.All
	require.NoError(t, spfViper.UnmarshalExact(&conf))
	sites, err := normalizedSites(conf)
	require.NoError(t, err)
	assert.Equal(t, []string{"garage"}, sites[0].loadpoints)
	assert.Equal(t, []string{"parking"}, sites[1].loadpoints)
	assert.Equal(t, "home-grid", sites[0].config["meters"].(map[string]any)["grid"])
	assert.Equal(t, "home-root", sites[0].circuit)
	assert.Equal(t, "home-grid-tariff", sites[0].tariffs.Grid)
	assert.Equal(t, "custom", sites[0].hems.Type)
}

func TestNormalizedSitesLegacy(t *testing.T) {
	conf := globalconfig.All{Site: map[string]any{"title": "Home"}}

	sites, err := normalizedSites(conf)
	require.NoError(t, err)
	require.Len(t, sites, 1)
	assert.Equal(t, "default", sites[0].name)
	assert.Equal(t, "Home", sites[0].config["title"])
}

func TestNormalizedSitesAssignments(t *testing.T) {
	conf := globalconfig.All{
		Sites: []globalconfig.Site{
			{Name: "home", Loadpoints: []string{"garage"}, Other: map[string]any{"title": "Home"}},
			{Name: "office", Other: map[string]any{"title": "Office"}},
		},
		Loadpoints: []config.Named{
			{Name: "garage"},
			{Name: "parking", Other: map[string]any{"site": "office"}},
		},
	}

	sites, err := normalizedSites(conf)
	require.NoError(t, err)
	require.Len(t, sites, 2)
	assert.Equal(t, []string{"garage"}, sites[0].loadpoints)
	assert.Equal(t, []string{"parking"}, sites[1].loadpoints)
}

func TestNormalizedSitesDatabaseLoadpointAssignment(t *testing.T) {
	conf := globalconfig.All{Sites: []globalconfig.Site{{Name: "home"}, {Name: "office"}}}
	databaseLoadpoint := config.Named{Name: "db:42", Other: map[string]any{"site": "office"}}

	sites, err := normalizedSites(conf, databaseLoadpoint)
	require.NoError(t, err)
	assert.Empty(t, sites[0].loadpoints)
	assert.Equal(t, []string{"db:42"}, sites[1].loadpoints)
}

func TestNormalizedSitesRejectsAmbiguousAssignments(t *testing.T) {
	conf := globalconfig.All{
		Sites: []globalconfig.Site{
			{Name: "home", Loadpoints: []string{"garage"}},
			{Name: "office", Loadpoints: []string{"garage"}},
		},
		Loadpoints: []config.Named{{Name: "garage"}},
	}

	_, err := normalizedSites(conf)
	require.EqualError(t, err, `loadpoint "garage" is assigned to sites "home" and "office"`)
}

func TestNormalizedSitesRejectsMixedSyntax(t *testing.T) {
	conf := globalconfig.All{
		Site:  map[string]any{"title": "Home"},
		Sites: []globalconfig.Site{{Name: "home"}},
	}

	_, err := normalizedSites(conf)
	require.EqualError(t, err, "site and sites cannot be used together")
}

func TestSitesYamlRejectsVehicleAssignments(t *testing.T) {
	spfViper.Reset()
	t.Cleanup(spfViper.Reset)
	spfViper.SetConfigType("yaml")
	require.NoError(t, spfViper.ReadConfig(strings.NewReader(`
sites:
  - name: home
    vehicles: [car]
`)))

	var conf globalconfig.All
	require.NoError(t, spfViper.UnmarshalExact(&conf))
	_, err := normalizedSites(conf)
	require.EqualError(t, err, `site "home" cannot assign global vehicles`)
}
