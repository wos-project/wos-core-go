package models

import (
	"gorm.io/gorm"
)

type Place struct {
	gorm.Model

	Name       string   `gorm:"column:name; binding:"required"`
	Location   PointGeo `gorm:"column:location"`
	ViewportNE PointGeo `gorm:"column:viewport_ne"`
	ViewportSW PointGeo `gorm:"column:viewport_sw"`
	Uid        string   `gorm:"column:uid; binding:"required"`
	Status     byte     `gorm:"column:status; index"`
	Metadata   string   `gorm:"column:metadata; binding:"required"`
}

type PlaceJson struct {
	Name     string `json:"name"`
	Geometry struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"location"`
		Viewport struct {
			Northeast struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"northeast,omitempty"`
			Southwest struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"southwest,omitempty"`
		} `json:"viewport,omitempty"`
	} `json:"geometry,omitempty"`

	Metadata  string `json:"metadata,omitempty"`
	Uid       string `json:"uid,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// MarshallToJson marshalls Place to PlaceJson
func (p *Place) MarshallToJson(includeMetadata bool) PlaceJson {
	var j PlaceJson
	j.Name = p.Name
	j.Uid = p.Uid
	j.CreatedAt = p.CreatedAt.String()
	j.Geometry.Location.Lat = p.Location.Lat
	j.Geometry.Location.Lon = p.Location.Lon
	j.Geometry.Viewport.Northeast.Lat = p.ViewportNE.Lat
	j.Geometry.Viewport.Northeast.Lon = p.ViewportNE.Lon
	j.Geometry.Viewport.Southwest.Lat = p.ViewportSW.Lat
	j.Geometry.Viewport.Southwest.Lon = p.ViewportSW.Lon

	if includeMetadata {
		j.Metadata = p.Metadata
	}
	return j
}

// MarshallFromJson marshalls Place to PlaceJson
func (p *Place) MarshallFromJson(j PlaceJson) {
	p.Name = j.Name
	p.Metadata = j.Metadata
	p.Location.Lat = j.Geometry.Location.Lat
	p.Location.Lon = j.Geometry.Location.Lon
	p.ViewportNE.Lat = j.Geometry.Viewport.Northeast.Lat
	p.ViewportNE.Lon = j.Geometry.Viewport.Northeast.Lon
	p.ViewportSW.Lat = j.Geometry.Viewport.Southwest.Lat
	p.ViewportSW.Lon = j.Geometry.Viewport.Southwest.Lon
}
