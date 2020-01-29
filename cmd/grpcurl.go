package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func invokeRPC(ctx context.Context, serverName, methodName string, protoFiles, importPath multiString, reqData []byte, res io.Writer) error {
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

	if len(protoFiles) > 0 {
		var err error
		descSource, err = grpcurl.DescriptorSourceFromProtoFiles(importPath, protoFiles...)
		if err != nil {
			return fmt.Errorf("create desc src: %w", err)
		}
	} else {
		md := grpcurl.MetadataFromHeaders([]string{})
		refCtx := metadata.NewOutgoingContext(ctx, md)
		cc = dial()
		refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))
		descSource = grpcurl.DescriptorSourceFromServer(ctx, refClient)
	}

	reset := func() {
		if refClient != nil {
			refClient.Reset()
			refClient = nil
		}
		if cc != nil {
			cc.Close()
			cc = nil
		}
	}
	defer reset()

	if cc == nil {
		cc = dial()
	}

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

type multiString []string

func (s *multiString) String() string {
	return strings.Join(*s, ",")
}

func (s *multiString) Set(value string) error {
	*s = append(*s, value)
	return nil
}
