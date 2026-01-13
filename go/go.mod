module github.com/leftbin/stigmer-sdk/go

go 1.24.3

toolchain go1.24.11

require github.com/leftbin/stigmer/apis/stubs/go v0.0.0-00010101000000-000000000000

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20251209175733-2a1774d88802.1 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

// TODO: Replace with Buf-generated manifest proto once Task 2 is complete
// For now, temporarily point to monorepo for development
replace github.com/leftbin/stigmer/apis/stubs/go => /Users/suresh/scm/github.com/leftbin/stigmer/apis/stubs/go/github.com/leftbin/stigmer/apis/stubs/go
