package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoHelper(t *testing.T) {
	lat1, lon1, lat2, lon2 := GeoBoundingBox(41.46777778, -71.29805556, 1)
	assert.Equal(t, float64(41.461418304120635), lat1)
	assert.Equal(t, float64(-71.30654287895551), lon1)
	assert.Equal(t, float64(41.4741366321536), lat2)
	assert.Equal(t, float64(-71.28956990580426), lon2)

	assert.Equal(t, 111.19492664455873, GeoDistance(90, 40, 91, 40))
	assert.Equal(t, 8.440664601894275, GeoDistance(41.473843, -71.245602, 41.495478, -71.342728))
}
