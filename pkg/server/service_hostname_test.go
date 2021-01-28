package server

import (
	"context"
	"strings"
	"testing"

	empty "github.com/golang/protobuf/ptypes/empty"
	grpctoken "github.com/hashicorp/horizon/pkg/grpc/token"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

func TestServiceRegisterHostname(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid auth", func(t *testing.T) {
		require := require.New(t)

		data := TestServer(t)
		client := data.Client

		// Get a hostname
		resp, err := client.RegisterHostname(ctx, &pb.RegisterHostnameRequest{
			Labels: &pb.LabelSet{
				Labels: []*pb.Label{
					{Name: "app", Value: "test"},
				},
			},
		}, grpc.PerRPCCredentials(grpctoken.Token("NOPE")))
		require.Error(err)
		require.Equal(codes.PermissionDenied, status.Code(err))
		require.Nil(resp)
	})

	t.Run("generated hostname", func(t *testing.T) {
		require := require.New(t)

		data := TestServer(t)
		client := data.Client
		optAuth := TestGuestAccount(t, client)

		// Should have no hostnames
		{
			resp, err := client.ListHostnames(ctx, &pb.ListHostnamesRequest{}, optAuth)
			require.NoError(err)
			require.NotNil(resp)
			require.Len(resp.Hostnames, 0)
		}

		// Get a hostname
		resp, err := client.RegisterHostname(ctx, &pb.RegisterHostnameRequest{
			Hostname: &pb.RegisterHostnameRequest_Generate{
				Generate: &empty.Empty{},
			},

			Labels: &pb.LabelSet{
				Labels: []*pb.Label{
					{Name: "app", Value: "test"},
				},
			},
		}, optAuth)
		require.NoError(err)
		require.NotNil(resp)
		require.NotEmpty(resp.Fqdn)

		// Should show up in the list
		{
			resp, err := client.ListHostnames(ctx, &pb.ListHostnamesRequest{}, optAuth)
			require.NoError(err)
			require.NotNil(resp)
			require.Len(resp.Hostnames, 1)
		}
	})

	t.Run("exact hostname", func(t *testing.T) {
		require := require.New(t)

		data := TestServer(t)
		client := data.Client
		optAuth := TestGuestAccount(t, client)

		// Get a hostname
		resp, err := client.RegisterHostname(ctx, &pb.RegisterHostnameRequest{
			Hostname: &pb.RegisterHostnameRequest_Exact{
				Exact: "foo",
			},

			Labels: &pb.LabelSet{
				Labels: []*pb.Label{
					{Name: "app", Value: "test"},
				},
			},
		}, optAuth)
		require.NoError(err)
		require.NotNil(resp)
		require.NotEmpty(resp.Fqdn)
		require.True(strings.HasPrefix(resp.Fqdn, "foo."))

		// Should be able to delete
		{
			resp, err := client.DeleteHostname(ctx, &pb.DeleteHostnameRequest{
				Hostname: "foo",
			}, optAuth)
			require.NoError(err)
			require.NotNil(resp)
		}

		// Should have no hostnames
		{
			resp, err := client.ListHostnames(ctx, &pb.ListHostnamesRequest{}, optAuth)
			require.NoError(err)
			require.NotNil(resp)
			require.Len(resp.Hostnames, 0)
		}
	})

	t.Run("rejects hostnames with double hyphen", func(t *testing.T) {
		require := require.New(t)

		data := TestServer(t)
		client := data.Client
		optAuth := TestGuestAccount(t, client)

		// Get a hostname
		_, err := client.RegisterHostname(ctx, &pb.RegisterHostnameRequest{
			Hostname: &pb.RegisterHostnameRequest_Exact{
				Exact: "foo--bar",
			},

			Labels: &pb.LabelSet{
				Labels: []*pb.Label{
					{Name: "app", Value: "test"},
				},
			},
		}, optAuth)
		require.Error(err)
	})

	t.Run("context cancellation doesn't dangle hostname", func(t *testing.T) {
		require := require.New(t)

		data := TestServer(t)
		client := data.Client
		optAuth := TestGuestAccount(t, client)

		// Should have no hostnames
		{
			resp, err := client.ListHostnames(ctx, &pb.ListHostnamesRequest{}, optAuth)
			require.NoError(err)
			require.NotNil(resp)
			require.Len(resp.Hostnames, 0)
		}

		// Create an expired context
		testLabelLinkFailed = true
		defer func() { testLabelLinkFailed = false }()

		// Get a hostname
		_, err := client.RegisterHostname(ctx, &pb.RegisterHostnameRequest{
			Hostname: &pb.RegisterHostnameRequest_Generate{
				Generate: &empty.Empty{},
			},

			Labels: &pb.LabelSet{
				Labels: []*pb.Label{
					{Name: "app", Value: "test"},
				},
			},
		}, optAuth)
		require.Error(err)

		// Should not show up in the list
		{
			resp, err := client.ListHostnames(ctx, &pb.ListHostnamesRequest{}, optAuth)
			require.NoError(err)
			require.NotNil(resp)
			require.Len(resp.Hostnames, 0)
		}
	})
}
