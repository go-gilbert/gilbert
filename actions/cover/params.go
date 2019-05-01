package cover

// toolArgsPrefixSize is prefix args count for 'go tool cover' command
//
// go test -coverprofile=/tmp/cover ./services/foo ./services/bar./services/baz
const toolArgsPrefixSize = 2

const (
	sumByStatements = "statements"
	sumByPercent    = "percent"
)

type params struct {
	Threshold  float64  `mapstructure:"threshold"`
	SumBy      string   `mapstructure:"sumBy"`
	Report     bool     `mapstructure:"reportCoverage"`
	FullReport bool     `mapstructure:"fullReport"`
	Packages   []string `mapstructure:"packages"`
}

func newParams() params {
	return params{
		Threshold: 0.0,
		SumBy:     sumByStatements,
		Report:    false,
	}
}
