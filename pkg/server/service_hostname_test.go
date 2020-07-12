package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

func TestServiceRegisterHostname(t *testing.T) {
	ctx := context.Background()

	t.Run("generated hostname", func(t *testing.T) {
		require := require.New(t)

		data := TestServer(t)
		client := data.Client
		optAuth := TestGuestAccount(t, client)

		// Get a hostname
		resp, err := client.RegisterHostname(ctx, &pb.RegisterHostnameRequest{
			Labels: &pb.LabelSet{
				Labels: []*pb.Label{
					{Name: "app", Value: "test"},
				},
			},
		}, optAuth)
		require.NoError(err)
		require.NotNil(resp)
		require.NotEmpty(resp.Fqdn)
	})
}
