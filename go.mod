module github.com/go-gilbert/gilbert

go 1.12

require (
	github.com/axw/gocov v0.0.0-20170322000131-3a69a0d2a4ef
	github.com/fatih/color v1.7.0
	github.com/go-gilbert/gilbert-sdk v0.9.0
	github.com/google/go-github/v25 v25.0.2
	github.com/hashicorp/hcl/v2 v2.1.0
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/rjeczalik/notify v0.9.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.20.0
	github.com/zclconf/go-cty v1.2.0
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/tools v0.0.0-20190420181800-aa740d480789
	gopkg.in/cheggaaa/pb.v1 v1.0.28
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/hashicorp/hcl/v2 => github.com/go-gilbert/hcl/v2 v2.3.1-0.20200316142931-bb845960b866
