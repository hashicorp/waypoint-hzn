// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pb

//go:generate sh -c "protoc -I../../proto --go_out=plugins=grpc:. --validate_out=\"lang=go:.\" ../../proto/*.proto"
