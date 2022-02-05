package models

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	migrations = []*gormigrate.Migration{
		{
			ID: "20220108000001",
			Migrate: func(tx *gorm.DB) error {
				err := tx.AutoMigrate(
					&Arc{},
					&Pin{},
					&PinnedArc{},
					&MediaUpload{},
					&Place{},
					&Layer{},
				)
				return err
			},
		},
	}

	// Db is the global database reference
	Db *gorm.DB

	tables = []string{
		"arcs",
		"pins",
		"pinned_arcs",
		"media_upload",
		"place",
		"layer",
	}
)

// InitializeDatabase opens and migrates the database if necessary
func InitializeDatabase() {
	if Db == nil {
		OpenDatabase()
	}

	m := gormigrate.New(Db, gormigrate.DefaultOptions, migrations)
	err := m.Migrate()
	if err != nil {
		glog.Fatalf("error migrating database %v", err)
	}
	glog.Info("migrated database")
}

// OpenDatabase opens the database and returns a handle
func OpenDatabase() {
	if Db == nil {
		dbinfo := viper.GetStringMap("database.main")
		if dbinfo["kind"].(string) != "postgres" {
			glog.Fatal("only postgres supported at this time")
		}
		dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
			dbinfo["username"].(string),
			dbinfo["password"].(string),
			dbinfo["database"].(string),
			dbinfo["host"].(string),
			dbinfo["port"].(int),
			dbinfo["sslmode"].(string))
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			glog.Fatalf("cannot open database %v", err)
		}
		Db = db
	}
}

// DropAllTables drops all tables and relations, only works with local database in debug mode
func DropAllTables() {
	dbinfo := viper.GetStringMap("database.main")
	if viper.GetString("mode") == "debug" && dbinfo["host"].(string) == "localhost" {
		for i := len(tables) - 1; i >= 0; i-- {
			t := tables[i]
			err := Db.Exec("DROP TABLE IF EXISTS " + t).Error
			if err != nil {
				glog.Fatalf("cannot drop tables %s %v", t, err)
			}
		}
		err := Db.Exec("DROP TABLE IF EXISTS migrations").Error
		if err != nil {
			glog.Fatalf("cannot drop table %v", err)
		}
	} else {
		glog.Warning("Not deleting tables since not in attached to local database and not in debug mode")
	}
}

// CloseDatabase closes the database
func CloseDatabase() {
	Db = nil
}

