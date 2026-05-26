package manifest

import (
	"fmt"
	"time"

	"github.com/go-viper/mapstructure/v2"
)

//var (
//	_ expr.Expression = (*Expression)(nil)
//)

// Period is job period in milliseconds
type Period uint

// ToDuration returns value in milliseconds for time.Duration
func (d Period) ToDuration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

// ActionParams is action params container
type ActionParams map[string]interface{}

// Unmarshal extracts action params into provided structure
func (p ActionParams) Unmarshal(dest interface{}) error {
	if err := mapstructure.Decode(p, dest); err != nil {
		return fmt.Errorf("failed to unmarshal action params, %w", err)
	}

	return nil
}
