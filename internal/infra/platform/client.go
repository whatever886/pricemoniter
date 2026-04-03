package platform

import (
	"context"
	"time"

	"kbfood/internal/domain/entity"
)

// Client represents a platform API client
type Client interface {
	// Name returns the platform name
	Name() string

	// FetchProducts fetches products from the platform (active mode)
	FetchProducts(ctx context.Context, region string) ([]*entity.PlatformProductDTO, error)

	// ShouldFetch checks if fetching is allowed at this time
	ShouldFetch(now time.Time) bool
}

// PushClient represents a platform that pushes data (passive mode)
type PushClient interface {
	// ProcessPushData processes pushed data from the platform
	ProcessPushData(ctx context.Context, data []*PushData) (map[string][]*entity.PlatformProductDTO, error)
}

// PushData represents pushed data from DT platform
type PushData struct {
	Title      string
	Price      float64
	OriginalPrice float64
	Status     int
	CrawlTime  int64
	Region     string
}

// RegionConfig holds latitude/longitude for regions
type RegionConfig struct {
	Name      string
	CityName  string
	Latitude  float64
	Longitude float64
}

// Default regions
var Regions = map[string]RegionConfig{
	"长沙": {
		Name:      "长沙",
		CityName:  "长沙市",
		Latitude:  28.2282,
		Longitude: 112.9388,
	},
	"东莞": {
		Name:      "东莞",
		CityName:  "东莞市",
		Latitude:  23.0207,
		Longitude: 113.7518,
	},
}

// GetRegionConfig returns region config, defaults to Changsha.
func GetRegionConfig(region string) RegionConfig {
	if cfg, ok := Regions[region]; ok {
		return cfg
	}
	return Regions["长沙"]
}

// ConvertToCityName converts region to city name
func ConvertToCityName(region string) string {
	cfg := GetRegionConfig(region)
	return cfg.CityName
}
