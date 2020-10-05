package image

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/stackrox/rox/roxctl/common/flags"
	"github.com/stackrox/rox/roxctl/image/check"
	"github.com/stackrox/rox/roxctl/image/scan"
)

// Command defines the image command tree
func Command() *cobra.Command {
	c := &cobra.Command{
		Use: "image",
	}

	c.AddCommand(check.Command())
	c.AddCommand(scan.Command())

	// This is set very high, because typically the scan will need to be triggered as the image will be new
	// This means we must let the scanners do their thing otherwise we will miss the scans
	// TODO(cgorman) We need a flag currently that says --wait-for-image timeout or something like that because Clair does scanning inline
	// but other scanners do not
	flags.AddTimeoutWithDefault(c, 10*time.Minute)
	return c
}
