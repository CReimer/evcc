package coordinator

import (
	"testing"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/loadpoint"
	"github.com/evcc-io/evcc/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCoordinatorSharesVehicleOwnership(t *testing.T) {
	ctrl := gomock.NewController(t)
	vehicle := api.NewMockVehicle(ctrl)
	home := loadpoint.NewMockAPI(ctrl)
	office := loadpoint.NewMockAPI(ctrl)

	coordinator := New(util.NewLogger("test"), []api.Vehicle{vehicle})
	homeVehicles := NewAdapter(home, coordinator)
	officeVehicles := NewAdapter(office, coordinator)

	require.Equal(t, []api.Vehicle{vehicle}, homeVehicles.GetVehicles(true))
	require.Equal(t, []api.Vehicle{vehicle}, officeVehicles.GetVehicles(true))

	homeVehicles.Acquire(vehicle)
	assert.Same(t, home, officeVehicles.Owner(vehicle))
	assert.Empty(t, officeVehicles.GetVehicles(true))

	home.EXPECT().SetVehicle(nil)
	officeVehicles.Acquire(vehicle)
	assert.Same(t, office, homeVehicles.Owner(vehicle))
	assert.Empty(t, homeVehicles.GetVehicles(true))
}

func TestCoordinatorAddIsIdempotent(t *testing.T) {
	ctrl := gomock.NewController(t)
	vehicle := api.NewMockVehicle(ctrl)
	coordinator := New(util.NewLogger("test"), []api.Vehicle{vehicle})

	coordinator.Add(vehicle)
	assert.Equal(t, []api.Vehicle{vehicle}, coordinator.GetVehicles(false))
}

func TestVehicleDetectByStatus(t *testing.T) {
	ctrl := gomock.NewController(t)

	type vehicle struct {
		*api.MockVehicle
		*api.MockChargeState
	}

	v1 := &vehicle{api.NewMockVehicle(ctrl), api.NewMockChargeState(ctrl)}
	v2 := &vehicle{api.NewMockVehicle(ctrl), api.NewMockChargeState(ctrl)}

	type testcase struct {
		string
		v1, v2 api.ChargeStatus
		res    api.Vehicle
	}
	tc := []testcase{
		{"A/A->0", api.StatusA, api.StatusA, nil},
		{"B/A->1", api.StatusB, api.StatusA, v1},
		{"B/A->1", api.StatusB, api.StatusA, v1},
		{"A/B->2", api.StatusA, api.StatusB, v2},
		{"A/B->2", api.StatusA, api.StatusB, v2},
		{"A/C->2", api.StatusA, api.StatusC, v2},
		{"A/C->2", api.StatusA, api.StatusC, v2},
		{"B/B->1", api.StatusB, api.StatusB, nil},
		{"B/C->1", api.StatusB, api.StatusC, v1},
		{"B/C->1", api.StatusB, api.StatusC, v1},
		{"C/B->2", api.StatusC, api.StatusB, v2},
		{"C/B->2", api.StatusC, api.StatusB, v2},
		{"C/C->1", api.StatusC, api.StatusC, nil},
	}

	log := util.NewLogger("foo")
	vehicles := []api.Vehicle{v1, v2}

	v1.MockVehicle.EXPECT().GetTitle().Return("v1").AnyTimes()
	v2.MockVehicle.EXPECT().GetTitle().Return("v2").AnyTimes()
	v1.MockVehicle.EXPECT().Identifiers().Return(nil).AnyTimes()
	v2.MockVehicle.EXPECT().Identifiers().Return([]string{"it's me"}).AnyTimes()
	v1.MockVehicle.EXPECT().Features().Return(nil).AnyTimes()
	v2.MockVehicle.EXPECT().Features().Return(nil).AnyTimes()

	var lp loadpoint.API
	c := New(log, vehicles)

	for _, tc := range tc {
		t.Logf("%+v", tc)

		v1.MockChargeState.EXPECT().Status().Return(tc.v1, nil)
		v2.MockChargeState.EXPECT().Status().Return(tc.v2, nil)

		available := c.availableDetectibleVehicles(lp) // include id-able vehicles
		res := c.identifyVehicleByStatus(available, api.StatusB)
		if tc.res != res {
			t.Errorf("expected %v, got %v", tc.res, res)
		}

		if res != nil {
			c.acquire(lp, res)
		} else {
			c.release(res)
		}
	}
}
