package metrics

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/evcc-io/evcc/db"
	"github.com/evcc-io/evcc/tariff"
	"gorm.io/gorm"
)

type meter struct {
	Meter        int      `json:"meter" gorm:"column:meter;uniqueIndex:meters_meter_ts"`
	Timestamp    int64    `json:"ts" gorm:"column:ts;uniqueIndex:meters_meter_ts"` // start of 15min slot
	Entity       entity   `json:"-" gorm:"foreignkey:Meter;references:Id"`
	Energy       float64  `json:"energy" gorm:"column:energy"`
	ReturnEnergy float64  `json:"returnEnergy" gorm:"column:return_energy"`
	SocTemp      *float64 `json:"socTemp,omitempty" gorm:"column:soc_temp"`    // at start of slot
	Recovered    bool     `json:"recovered,omitempty" gorm:"column:recovered"` // downtime catchup slot, excluded from profile
}

type entity struct {
	Id                int      `gorm:"column:id;primarykey"`
	Site              string   `gorm:"column:site;default:'';uniqueIndex:entities_site_group_name"`
	Group             string   `gorm:"column:group;uniqueIndex:entities_site_group_name"`
	Name              string   `gorm:"column:name;uniqueIndex:entities_site_group_name"`
	Title             string   `gorm:"column:title"`
	IsTemp            bool     `gorm:"column:is_temp"`             // soc_temp holds temperature, not soc
	EnergyMeter       *float64 `gorm:"column:energy_meter"`        // kWh, at last persisted slot
	ReturnEnergyMeter *float64 `gorm:"column:return_energy_meter"` // kWh, at last persisted slot
}

func init() {
	db.Register(func(_ *gorm.DB) error {
		return SetupSchema()
	})
}

// SetupSchema is used for testing
func SetupSchema() error {
	m := db.Instance.Migrator()

	// entities: add the scope before replacing the legacy unique index
	if m.HasTable(new(entity)) && !m.HasColumn(new(entity), "site") {
		if err := m.AddColumn(new(entity), "Site"); err != nil {
			return err
		}
	}
	if m.HasTable(new(entity)) {
		if err := db.Instance.Exec("UPDATE entities SET site = '' WHERE site IS NULL").Error; err != nil {
			return err
		}
	}
	for _, name := range []string{"entities_group_name", "idx_entities_group_name"} {
		if m.HasIndex(new(entity), name) {
			if err := m.DropIndex(new(entity), name); err != nil {
				return err
			}
		}
	}
	if err := db.Instance.AutoMigrate(new(entity)); err != nil {
		return err
	}

	// ensure home entity exists (reserves id=1 for legacy meter FK references)
	if _, err := createEntity(Home, Home, Home); err != nil {
		return err
	}

	// grid: migrate old entities (title was the device ref) to the fixed title
	if err := db.Instance.Model(new(entity)).Where(`"group" = ? AND title <> ?`, Grid, Grid).Update("title", Grid).Error; err != nil {
		return err
	}

	// enable FK constraints only here to make sure entity for metric exists
	if err := db.Instance.Exec("pragma foreign_keys(1)").Error; err != nil {
		return err
	}

	// drop obsolete indexes
	for _, idx := range []struct {
		name string
		obj  any
	}{
		{"name_idx", new(entity)},
		{"group_name", new(entity)},
		{"meter_ts", new(meter)},
	} {
		if m.HasIndex(idx.obj, idx.name) {
			if err := m.DropIndex(idx.obj, idx.name); err != nil {
				return err
			}
		}
	}

	rename := func(from, to string) error {
		if table := new(meter); m.HasColumn(table, from) && !m.HasColumn(table, to) {
			return m.RenameColumn(table, from, to)
		}
		return nil
	}

	// meter: split energy direction
	if err := rename("val", "import"); err != nil {
		return err
	}

	// meter: split energy direction #2
	if err := rename("pos", "import"); err != nil {
		return err
	}
	if err := rename("neg", "export"); err != nil {
		return err
	}

	// meter: rename to energy/return_energy
	if err := rename("import", "energy"); err != nil {
		return err
	}
	if err := rename("export", "return_energy"); err != nil {
		return err
	}

	// meter: ts migration
	if m.HasTable(new(meter)) {
		types, err := m.ColumnTypes(new(meter))
		if err != nil {
			return err
		}
		tsIdx := slices.IndexFunc(types, func(typ gorm.ColumnType) bool {
			return typ.Name() == "ts"
		})
		if tsIdx == -1 {
			return errors.New("missing meters.ts")
		}

		if tsTyp, _ := types[tsIdx].ColumnType(); !strings.EqualFold(tsTyp, "INTEGER") {
			db, err := db.Instance.DB()
			if err != nil {
				return err
			}

			if _, err := db.Exec(`UPDATE meters SET ts = unixepoch(ts)`); err != nil {
				return err
			}
		}
	}

	return db.Instance.AutoMigrate(new(meter))
}

var persistSubscribers struct {
	sync.RWMutex
	callbacks []func(time.Time)
}

// SubscribePersist adds a callback invoked after a completed slot is written.
func SubscribePersist(callback func(time.Time)) {
	persistSubscribers.Lock()
	defer persistSubscribers.Unlock()
	persistSubscribers.callbacks = append(persistSubscribers.callbacks, callback)
}

func notifyPersist(slot time.Time) {
	persistSubscribers.RLock()
	callbacks := slices.Clone(persistSubscribers.callbacks)
	persistSubscribers.RUnlock()
	for _, callback := range callbacks {
		callback(slot)
	}
}

// persist stores a completed 15min slot
func persist(entity entity, ts time.Time, energy, returnEnergy float64, socTemp *float64, recovered bool) error {
	slot := ts.Truncate(tariff.SlotDuration)
	if err := db.Instance.Create(&meter{
		Meter:        entity.Id,
		Timestamp:    slot.Unix(),
		Energy:       energy,
		ReturnEnergy: returnEnergy,
		SocTemp:      socTemp,
		Recovered:    recovered,
	}).Error; err != nil {
		return err
	}
	notifyPersist(slot)
	return nil
}
