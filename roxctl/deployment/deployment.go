package deployment

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/stackrox/rox/roxctl/common/flags"
	"github.com/stackrox/rox/roxctl/deployment/check"
)

// Command defines the image command tree
func Command() *cobra.Command {
	c := &cobra.Command{
		Use: "deployment",
	}

	c.AddCommand(check.Command())
	flags.AddTimeoutWithDefault(c, 1*time.Minute)
	return c
}
