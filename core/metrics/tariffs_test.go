package metrics

import (
	"testing"
	"time"

	"github.com/evcc-io/evcc/db"
	"github.com/stretchr/testify/require"
)

func TestPersistTariffs(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, db.Instance.AutoMigrate(new(tariffValue)))

	slot := time.Date(2026, 4, 15, 16, 15, 0, 0, time.UTC)
	grid, co2 := 0.3, 250.0

	// nil values omitted
	require.NoError(t, PersistTariffs(slot, &grid, nil, &co2, nil))

	var res tariffValue
	require.NoError(t, db.Instance.First(&res).Error)
	require.Equal(t, slot.Unix(), res.Timestamp)
	require.InDelta(t, 0.3, *res.Grid, 0.001)
	require.Nil(t, res.FeedIn)
	require.InDelta(t, 250, *res.Co2, 0.001)
	require.Nil(t, res.Temperature)

	// duplicate slot ignored, first values kept
	other := 0.4
	require.NoError(t, PersistTariffs(slot, &other, nil, nil, nil))

	var count int64
	require.NoError(t, db.Instance.Model(new(tariffValue)).Count(&count).Error)
	require.Equal(t, int64(1), count)

	require.NoError(t, db.Instance.First(&res).Error)
	require.InDelta(t, 0.3, *res.Grid, 0.001)

	// all nil: no row
	require.NoError(t, PersistTariffs(slot.Add(15*time.Minute), nil, nil, nil, nil))
	require.NoError(t, db.Instance.Model(new(tariffValue)).Count(&count).Error)
	require.Equal(t, int64(1), count)

	// all values set: each column mapped independently
	next := slot.Add(30 * time.Minute)
	feedin, temp := 0.08, 21.5
	require.NoError(t, PersistTariffs(next, &grid, &feedin, &co2, &temp))

	require.NoError(t, db.Instance.Where("ts = ?", next.Unix()).First(&res).Error)
	require.InDelta(t, 0.3, *res.Grid, 0.001)
	require.InDelta(t, 0.08, *res.FeedIn, 0.001)
	require.InDelta(t, 250, *res.Co2, 0.001)
	require.InDelta(t, 21.5, *res.Temperature, 0.001)
}

func TestPersistTariffsSiteScope(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, setupTariffSchema())

	slot := time.Now().Truncate(15 * time.Minute)
	home, office := 0.2, 0.4
	require.NoError(t, PersistSiteTariffs("home_1", slot, &home, nil, nil, nil))
	require.NoError(t, PersistSiteTariffs("homeA1", slot, &office, nil, nil, nil))

	var values []tariffValue
	require.NoError(t, db.Instance.Order("site").Find(&values).Error)
	require.Len(t, values, 2)
	require.InDelta(t, office, *values[0].Grid, 0.001)
	require.InDelta(t, home, *values[1].Grid, 0.001)
}

func TestTariffSchemaMigrationAddsSiteIdentity(t *testing.T) {
	type legacyTariff struct {
		Site      *string `gorm:"column:site"`
		Timestamp int64   `gorm:"column:ts;uniqueIndex"`
	}

	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, db.Instance.Exec("DROP TABLE tariffs").Error)
	require.NoError(t, db.Instance.Table("tariffs").AutoMigrate(new(legacyTariff)))
	require.NoError(t, db.Instance.Table("tariffs").Create(map[string]any{"ts": int64(1), "site": nil}).Error)
	require.NoError(t, setupTariffSchema())
	var nullSites int64
	require.NoError(t, db.Instance.Table("tariffs").Where("site IS NULL").Count(&nullSites).Error)
	require.Zero(t, nullSites)

	slot := time.Now().Truncate(15 * time.Minute)
	home, office := 0.2, 0.4
	require.NoError(t, PersistSiteTariffs("home", slot, &home, nil, nil, nil))
	require.NoError(t, PersistSiteTariffs("office", slot, &office, nil, nil, nil))
}
