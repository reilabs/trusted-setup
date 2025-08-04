// Package api contains definitions for the gRPC protocol between the ceremony client and ceremony server.
//
// The protocol definition is in ceremony.proto. To regenerate the Go code, run: `go generate ./...`.
//
// Below is the go-generate directive that will be invoked automatically on code regeneration. This is an input
// for the go compiler, please do not remove this comment.
//
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ceremony.proto
package api

// ProtocolVersion specifies the version of the protocol.
//
// The version must match between client and server, otherwise client's connection request will be rejected.
// Bump this version when ceremony.proto is changed.
const ProtocolVersion = 0x0001
