package hetznerdns

import (
	"context"
)

type DNS interface {
	CreateZone(context.Context, string, uint64) (Zone, error)
	GetZone(context.Context, ZoneID) (Zone, error)
	GetZones(context.Context) (Zones, error)
	GetRecords(context.Context, GetRecordsOpts) (Records, error)
}
