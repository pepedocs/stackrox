package environment

import (
	"github.com/stackrox/rox/roxctl/common"
	"google.golang.org/grpc"
)

// Environment abstracts the common.RoxctlHTTPClient, IO and grpc.ClientConn used within each command of the CLI.
//go:generate mockgen-wrapper
type Environment interface {
	// HTTPClient returns a common.RoxctlHTTPClient
	HTTPClient() common.RoxctlHTTPClient

	// GRPCConnection returns an authenticated grpc.ClientConn
	GRPCConnection() (*grpc.ClientConn, error)

	// InputOutput returns an IO which hols all input / output streams
	InputOutput() IO
}
