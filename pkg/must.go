package pkg

import (
	"log"

	"google.golang.org/api/sheets/v4"
)

func MustValueRange(vr *sheets.ValueRange, err error) *sheets.ValueRange {
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return vr
}

func MustService(ss *sheets.Service, err error) *sheets.Service {
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return ss
}
