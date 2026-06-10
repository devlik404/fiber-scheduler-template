package timezone

import (
	_ "embed"
	"fmt"
	"time"
)

const (
	AsiaJakarta = "Asia/Jakarta"
	Default     = AsiaJakarta
)

//go:embed Asia/Jakarta
var asiaJakartaData []byte

func Load(name string) (*time.Location, error) {
	if name == "" {
		name = Default
	}

	if name == AsiaJakarta {
		location, err := time.LoadLocationFromTZData(AsiaJakarta, asiaJakartaData)
		if err != nil {
			return nil, fmt.Errorf("load embedded timezone %q: %w", AsiaJakarta, err)
		}

		return location, nil
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", name, err)
	}

	return location, nil
}
