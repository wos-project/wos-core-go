package utils

import (
	"github.com/kellydunn/golang-geo"
	"math"
)

// GeoBoundingBox returns bounding box given center point and range d in kilometers
// This is an estimate since uses the hypotenuse at NW and SE
func GeoBoundingBox(lat float64, lon float64, d float64) (lat1 float64, lon1 float64, lat2 float64, lon2 float64) {

	p := geo.NewPoint(lat, lon)
	p1 := p.PointAtDistanceAndBearing(d, 135)
	p2 := p.PointAtDistanceAndBearing(d, 315)
	return math.Min(p1.Lat(), p2.Lat()), math.Min(p1.Lng(), p2.Lng()), math.Max(p1.Lat(), p2.Lat()), math.Max(p1.Lng(), p2.Lng())
}

// GeoDistance returns the greater circle distance in kilometers between two points
func GeoDistance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {

	p1 := geo.NewPoint(lat1, lon1)
	p2 := geo.NewPoint(lat2, lon2)
	return p1.GreatCircleDistance(p2)
}
