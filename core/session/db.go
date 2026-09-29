package session

import (
	"github.com/evcc-io/evcc/db"
	"github.com/evcc-io/evcc/util"
	"gorm.io/gorm"
)

// DB is a SQL database storage service
type DB struct {
	log  *util.Logger
	db   *gorm.DB
	site string
	name string
}

var sessions Sessions

func init() {
	db.Register(func(db *gorm.DB) error {
		if err := setupSchema(db); err != nil {
			return err
		}

		return db.Find(&sessions).Error
	})
}

func setupSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(new(Session)); err != nil {
		return err
	}
	return db.Exec("UPDATE sessions SET site = '' WHERE site IS NULL").Error
}

// NewStore creates a session store
func NewStore(name string, db *gorm.DB, site ...string) (*DB, error) {
	err := setupSchema(db)
	siteName := ""
	if len(site) > 0 {
		siteName = site[0]
	}
	if err == nil && siteName != "" {
		legacyName := siteName + "/" + name
		err = db.Model(new(Session)).Where("COALESCE(site, '') = '' AND loadpoint IN ?", []string{legacyName, name}).
			Updates(map[string]any{"site": siteName, "loadpoint": name}).Error
	}

	sessiondb := &DB{
		log:  util.NewLogger("db"),
		db:   db,
		site: siteName,
		name: name,
	}

	return sessiondb, err
}

// New creates a charging session
func (s *DB) New(meter float64) *Session {
	t := Session{
		Site:      s.site,
		Loadpoint: s.name,
	}

	if meter > 0 {
		t.MeterStart = &meter
	}

	return &t
}

// Persist creates or updates a transaction in the database
func (s *DB) Persist(session any) {
	if err := s.db.Save(session).Error; err != nil {
		s.log.ERROR.Printf("persist: %v", err)
	}
}

// Return sessions
// TODO make this part of db
func (s *DB) Sessions() (Sessions, error) {
	var res Sessions
	tx := s.db.Find(&res)
	return res, tx.Error
}

func (s *DB) ClosePendingSessionsInHistory(chargeMeterTotal float64) error {
	var res Sessions
	if tx := s.db.Find(&res, map[string]any{"finished": "0001-01-01 00:00:00+00:00", "site": s.site, "loadpoint": s.name}); tx.Error != nil {
		return tx.Error
	}

	for _, session := range res {
		var nextSession Session

		var tx *gorm.DB
		if tx = s.db.Limit(1).Order("ID").Find(&nextSession, "ID > ? AND site = ? AND loadpoint = ?", session.ID, s.site, s.name); tx.Error != nil {
			return tx.Error
		}

		if tx.RowsAffected == 0 {
			// no successor, this is the most recent session and it is open
			session.MeterStop = &chargeMeterTotal
		} else {
			session.MeterStop = nextSession.MeterStart
		}

		if session.MeterStart != nil && session.MeterStop != nil {
			session.ChargedEnergy = *session.MeterStop - *session.MeterStart
			s.Persist(session)
		}
	}

	return nil
}
