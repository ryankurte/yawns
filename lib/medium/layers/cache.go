package layers

import (
	"fmt"

	"github.com/ryankurte/yawns/lib/types"
)

// Cache is a simple map based cache to minimise computations required for each layer
type Cache struct {
	cache map[string]float64
}

// NewCache creates a cache for attenuation v
func NewCache() Cache {
	return Cache{make(map[string]float64)}
}

func (c *Cache) key(band types.Frequency, a, b types.Location) string {
	return fmt.Sprintf("%f %s %s", band, a, b)
}

// Set adds an attenuation value for a given band and node pair
func (c *Cache) Set(band types.Frequency, a, b types.Location, attenuation float64) {
	key := c.key(band, a, b)
	c.cache[key] = attenuation
}

// Get fetches an attenuation value (if available) for a given band and node pair
func (c *Cache) Get(band types.Frequency, a, b types.Location) (float64, bool) {
	key := c.key(band, a, b)
	v, ok := c.cache[key]
	return v, ok
}
