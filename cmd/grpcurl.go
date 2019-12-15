package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func invokeRPC(ctx context.Context, serverName, methodName string, reqData []byte, res io.Writer) error {
	dial := func() *grpc.ClientConn {
		var creds credentials.TransportCredentials
		var opts []grpc.DialOption
		network := "tcp"
		cc, err := grpcurl.BlockingDial(ctx, network, serverName, creds, opts...)
		if err != nil {
			log.Fatal(err)
		}

		return cc
	}

	var cc *grpc.ClientConn
	var descSource grpcurl.DescriptorSource
	var refClient *grpcreflect.Client

	md := grpcurl.MetadataFromHeaders([]string{})
	refCtx := metadata.NewOutgoingContext(ctx, md)
	cc = dial()
	refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))
	descSource = grpcurl.DescriptorSourceFromServer(ctx, refClient)

	var in io.Reader
	in = bytes.NewReader(reqData)

	rf, formatter, err := grpcurl.RequestParserAndFormatterFor(grpcurl.FormatJSON, descSource, false, false, in)
	if err != nil {
		return err
	}
	h := grpcurl.NewDefaultEventHandler(res, descSource, formatter, false)

	err = grpcurl.InvokeRPC(ctx, descSource, cc, methodName, []string{}, h, rf.Next)
	if err != nil {
		return err
	}

	if h.Status.Code() != codes.OK {
		grpcurl.PrintStatus(os.Stderr, h.Status, formatter)
		return fmt.Errorf("invalid response code: %s", h.Status.Code().String())
	}

	return nil
}
