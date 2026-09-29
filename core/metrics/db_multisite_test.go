package metrics

import (
	"testing"
	"time"

	"github.com/evcc-io/evcc/db"
	"github.com/stretchr/testify/require"
)

func TestNewSiteCollectorMigratesPrefixedEntity(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, SetupSchema())

	legacy := entity{Group: Grid, Name: "home_1/grid", Title: Grid}
	require.NoError(t, db.Instance.Create(&legacy).Error)
	previous := entity{Group: Grid, Name: Grid, Title: Grid}
	require.NoError(t, db.Instance.Create(&previous).Error)
	now := time.Now().Truncate(15 * time.Minute)
	require.NoError(t, db.Instance.Create(&[]meter{
		{Meter: previous.Id, Timestamp: now.Add(-15 * time.Minute).Unix(), Energy: 1.1},
		{Meter: legacy.Id, Timestamp: now.Unix(), Energy: 4.2},
	}).Error)

	collector, err := NewSiteCollector("home_1", Grid, Grid, Grid)
	require.NoError(t, err)
	require.Equal(t, legacy.Id, collector.entity.Id)
	require.Equal(t, "home_1", collector.entity.Site)
	require.Equal(t, Grid, collector.entity.Name)

	site := "home_1"
	series, err := QueryEnergy(time.Time{}, time.Time{}, "15m", false, EnergyFilter{Site: &site})
	require.NoError(t, err)
	require.Len(t, series, 1)
	require.Len(t, series[0].Data, 2)
	require.Equal(t, 1.1, series[0].Data[0].Energy)
	require.Equal(t, 4.2, series[0].Data[1].Energy)
	var count int64
	require.NoError(t, db.Instance.Model(new(entity)).Where(`"group" = ? AND name = ?`, Grid, Grid).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestEntityIdentityIncludesSite(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, SetupSchema())

	home, err := NewSiteCollector("home_1", Grid, Grid, Grid)
	require.NoError(t, err)
	office, err := NewSiteCollector("homeA1", Grid, Grid, Grid)
	require.NoError(t, err)
	require.NotEqual(t, home.entity.Id, office.entity.Id)
}

func TestSetupSchemaReplacesLegacyEntityIdentity(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, db.Instance.Exec("DROP TABLE meters").Error)
	require.NoError(t, db.Instance.Exec("DROP TABLE entities").Error)
	require.NoError(t, db.Instance.Exec(`CREATE TABLE entities (
		id integer primary key, "group" text, name text, title text, is_temp numeric,
		energy_meter real, return_energy_meter real, site text
	)`).Error)
	require.NoError(t, db.Instance.Exec(`CREATE UNIQUE INDEX entities_group_name ON entities ("group", name)`).Error)
	require.NoError(t, db.Instance.Exec(`INSERT INTO entities ("group", name, site) VALUES (?, ?, NULL)`, Grid, Grid).Error)
	require.NoError(t, SetupSchema())
	var nullSites int64
	require.NoError(t, db.Instance.Model(new(entity)).Where("site IS NULL").Count(&nullSites).Error)
	require.Zero(t, nullSites)

	_, err := NewSiteCollector("home", Grid, Grid, Grid)
	require.NoError(t, err)
	_, err = NewSiteCollector("office", Grid, Grid, Grid)
	require.NoError(t, err)
}
