package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PointGeo is a manually serialized postgis point
type PointGeo struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

// GormDataType sets the Gorm type to geometry
func (p *PointGeo) GormDataType() string {
	return "geometry(POINT,4326)"
}

// GormValue sets geom point
func (p PointGeo) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "ST_PointFromText(?, 4326)",
		Vars: []interface{}{fmt.Sprintf("POINT(%f %f)", p.Lon, p.Lat)},
	}
}

// String converts simple point to postgis string
func (p *PointGeo) String() string {
	return fmt.Sprintf("SRID=4326;POINT(%v %v)", p.Lon, p.Lat)
}

// Scan reads a point from database and converts to simple point
func (p *PointGeo) Scan(val interface{}) error {
	b, err := hex.DecodeString(val.(string))
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	var wkbByteOrder uint8
	if err := binary.Read(r, binary.LittleEndian, &wkbByteOrder); err != nil {
		return err
	}

	var byteOrder binary.ByteOrder
	switch wkbByteOrder {
	case 0:
		byteOrder = binary.BigEndian
	case 1:
		byteOrder = binary.LittleEndian
	default:
		return fmt.Errorf("Invalid byte order %d", wkbByteOrder)
	}

	var wkbGeometryType uint64
	if err := binary.Read(r, byteOrder, &wkbGeometryType); err != nil {
		return err
	}

	if err := binary.Read(r, byteOrder, p); err != nil {
		return err
	}

	return nil
}

// Value converts simple point to SQL driver value, which is a string
func (p PointGeo) Value() (driver.Value, error) {
	return p.String(), nil
}
