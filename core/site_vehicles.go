package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/keys"
	"github.com/evcc-io/evcc/core/site"
	"github.com/evcc-io/evcc/core/vehicle"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/config"
	"github.com/samber/lo"
)

type planStruct struct {
	Soc  int       `json:"soc"`
	Time time.Time `json:"time"`
}

type vehicleStruct struct {
	Title          string              `json:"title"`
	Icon           string              `json:"icon,omitempty"`
	Capacity       float64             `json:"capacity,omitempty"`
	Phases         int                 `json:"phases,omitempty"`
	Mode           api.ChargeMode      `json:"mode,omitempty"`
	AlwaysCharge   api.AlwaysCharge    `json:"alwaysCharge,omitempty"`
	MinSoc         int                 `json:"minSoc,omitempty"`
	LimitSoc       int                 `json:"limitSoc,omitempty"`
	MinCurrent     float64             `json:"minCurrent,omitempty"`
	MaxCurrent     float64             `json:"maxCurrent,omitempty"`
	Priority       int                 `json:"priority,omitempty"`
	Features       []string            `json:"features,omitempty"`
	Plan           *planStruct         `json:"plan,omitempty"`
	RepeatingPlans []api.RepeatingPlan `json:"repeatingPlans"`
	PlanStrategy   api.PlanStrategy    `json:"planStrategy"`
}

// publishVehicles returns a list of vehicle titles
func (site *Site) publishVehicles() {
	vv := site.Vehicles().Settings()
	res := make(map[string]vehicleStruct, len(vv))

	for _, v := range vv {
		instance := v.Instance()
		if instance == nil {
			continue
		}

		ac := instance.OnIdentified()

		var plan *planStruct
		if time, soc := v.GetPlanSoc(); !time.IsZero() {
			plan = &planStruct{
				Soc:  soc,
				Time: time,
			}
		}

		res[v.Name()] = vehicleStruct{
			Title:          instance.GetTitle(),
			Icon:           instance.Icon(),
			Capacity:       instance.Capacity(),
			Phases:         instance.Phases(),
			Mode:           v.GetMode(),
			AlwaysCharge:   v.GetAlwaysCharge(),
			MinSoc:         v.GetMinSoc(),
			LimitSoc:       v.GetLimitSoc(),
			MinCurrent:     ac.MinCurrent,
			MaxCurrent:     ac.MaxCurrent,
			Priority:       ac.Priority,
			Features:       lo.Map(instance.Features(), func(f api.Feature, _ int) string { return f.String() }),
			Plan:           plan,
			RepeatingPlans: v.GetRepeatingPlans(),
			PlanStrategy:   v.GetPlanStrategy(),
		}

		// publish effective plan strategy immediately for soc-based planning
		if lp := site.coordinator.Owner(instance); lp != nil {
			lp.PublishEffectiveValues()
		}
	}

	site.publish(keys.Vehicles, res)
}

// updateVehicles adds or removes a vehicle asynchronously
func (site *Site) updateVehicles(op config.Operation, dev config.Device[api.Vehicle]) {
	idx := slices.IndexFunc(site.vehicles, func(existing config.Device[api.Vehicle]) bool {
		return existing.Config().Name == dev.Config().Name
	})
	vehicle := dev.Instance()

	switch op {
	case config.OpAdd:
		if idx >= 0 {
			return
		}
		site.vehicles = append(site.vehicles, dev)
		site.coordinator.Add(vehicle)

	case config.OpDelete:
		if idx < 0 {
			return
		}
		site.vehicles = slices.Delete(site.vehicles, idx, idx+1)
		site.coordinator.Delete(vehicle)
	}

	// TODO remove vehicle from mqtt
	site.publishVehicles()
}

var _ site.Vehicles = (*vehicles)(nil)

type vehicles struct {
	log     *util.Logger
	devices []config.Device[api.Vehicle]
}

func (vv *vehicles) Instances() []api.Vehicle {
	return config.Instances(vv.devices)
}

func (vv *vehicles) Settings() []vehicle.API {
	res := make([]vehicle.API, 0, len(vv.devices))
	for _, dev := range vv.devices {
		// skip disabled vehicles
		if dev.Instance() == nil {
			continue
		}
		res = append(res, vehicle.Adapter(vv.log, dev))
	}

	return res
}

func (vv *vehicles) ByName(name string) (vehicle.API, error) {
	for _, dev := range vv.devices {
		if dev.Config().Name == name {
			return vehicle.Adapter(vv.log, dev), nil
		}
	}
	return nil, fmt.Errorf("vehicle not found: %s", name)
}
