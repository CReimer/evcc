package session

import (
	"testing"

	"github.com/evcc-io/evcc/db"
	"github.com/stretchr/testify/require"
)

func TestNewStoreMigratesPrefixedSessions(t *testing.T) {
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, db.Instance.Exec("DROP TABLE sessions").Error)
	require.NoError(t, db.Instance.Exec("CREATE TABLE sessions (id integer primary key, loadpoint text, charged_kwh real, site text)").Error)
	require.NoError(t, db.Instance.Exec("INSERT INTO sessions (loadpoint, charged_kwh, site) VALUES (?, ?, NULL), (?, ?, NULL)", "garage", 1.2, "home_1/garage", 12.3).Error)

	store, err := NewStore("garage", db.Instance, "home_1")
	require.NoError(t, err)

	var got Sessions
	require.NoError(t, db.Instance.Order("id").Find(&got).Error)
	require.Len(t, got, 2)
	for _, item := range got {
		require.Equal(t, "home_1", item.Site)
		require.Equal(t, "garage", item.Loadpoint)
	}
	require.Equal(t, "home_1", store.New(0).Site)
}
