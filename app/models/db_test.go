package models

import (
	"testing"

	"gorm.io/datatypes"
	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.SetConfigFile("../config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		glog.Fatalf("Error initializing config, %s", err)
	}
}

func TestDatabase(t *testing.T) {
	OpenDatabase()
	DropAllTables()
	Db.Exec("DROP TABLE IF EXISTS migrations")
	CloseDatabase()

	InitializeDatabase()

	// create sample data
	pt := PointGeo{Lat: 42.1, Lon: -72.3}
	specs := map[string]interface{}{
		"name": "gaia",
		"fidelity": []string{"phone", "glasses"},
		"pin": map[string]interface{}{
		  "point": []float64{41.5, -71.3},
		},
	      }
	w := Pin{
		Object: Object{
			Cid: "xxxyyy", 
			Name: "test", 
			Body: specs,
		},
		Location: pt,
	}
	
	res := Db.Create(&w)
	assert.Nil(t, res.Error)

	var count int
	// distance in meters, so 10 meters here
	res = Db.Raw("SELECT COUNT(*) FROM pins WHERE ST_DistanceSphere('SRID=4326;POINT(-73.1235 42.3521)'::geometry, location) < 10").Scan(&count)
	assert.Nil(t, res.Error)
	assert.Equal(t, 0, count)

	var dist float64
	res = Db.Raw("SELECT ST_DistanceSphere('SRID=4326;POINT(-73.3 42.1)'::geometry, location) FROM pins").Scan(&dist)
	assert.Nil(t, res.Error)
	assert.Equal(t, float64(82503.59211615), dist)

	var w2 Pin
	res = Db.Raw("SELECT * FROM pins WHERE ST_DistanceSphere('SRID=4326;POINT(-73.1235 42.3521)'::geometry, location) < 100000").Scan(&w2)
	assert.Nil(t, res.Error)
	assert.Equal(t, float64(42.1), w2.Location.Lat)
	assert.Equal(t, float64(-72.3), w2.Location.Lon)
	assert.Equal(t, JSONMap(JSONMap{"fidelity":[]interface {}{"phone", "glasses"}, "name":"gaia", "pin":map[string]interface {}{"point":[]interface {}{41.5, -71.3}}}), w2.Object.Body)

	// test JSONB
	w3 := Pin{}
  	Db.First(&w3, datatypes.JSONQuery("body").HasKey("name", "gaia"))
	assert.Equal(t, "test", w3.Name)
}
