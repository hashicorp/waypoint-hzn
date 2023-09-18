// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/waypoint-hzn/pkg/pb"
)

func TestServiceRegisterGuestAccount(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	data := TestServer(t)
	client := data.Client

	resp, err := client.RegisterGuestAccount(ctx, &pb.RegisterGuestAccountRequest{
		ServerId:  "A",
		AcceptTos: true,
	})
	require.NoError(err)
	require.NotNil(resp)
	require.NotEmpty(resp.Token)
}
