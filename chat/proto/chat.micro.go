// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: chat/proto/chat.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	chat/proto/chat.proto

It has these top-level messages:
	SendPMReq
	SendPMRes
	PushMsg
	SentPrivateMsg
	UserOnlineReq
	UserOnlineRes
	UserOfflineReq
	UserOfflineRes
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
	context "context"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for Private service

type PrivateService interface {
	Send(ctx context.Context, in *SendPMReq, opts ...client.CallOption) (*SendPMRes, error)
}

type privateService struct {
	c           client.Client
	serviceName string
}

func PrivateServiceClient(serviceName string, c client.Client) PrivateService {
	if c == nil {
		c = client.NewClient()
	}
	if len(serviceName) == 0 {
		serviceName = "proto"
	}
	return &privateService{
		c:           c,
		serviceName: serviceName,
	}
}

func (c *privateService) Send(ctx context.Context, in *SendPMReq, opts ...client.CallOption) (*SendPMRes, error) {
	req := c.c.NewRequest(c.serviceName, "Private.Send", in)
	out := new(SendPMRes)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Private service

type PrivateHandler interface {
	Send(context.Context, *SendPMReq, *SendPMRes) error
}

func RegisterPrivateHandler(s server.Server, hdlr PrivateHandler, opts ...server.HandlerOption) {
	s.Handle(s.NewHandler(&Private{hdlr}, opts...))
}

type Private struct {
	PrivateHandler
}

func (h *Private) Send(ctx context.Context, in *SendPMReq, out *SendPMRes) error {
	return h.PrivateHandler.Send(ctx, in, out)
}

// Client API for State service

type StateService interface {
	Online(ctx context.Context, in *UserOnlineReq, opts ...client.CallOption) (*UserOnlineRes, error)
	Offline(ctx context.Context, in *UserOfflineReq, opts ...client.CallOption) (*UserOfflineRes, error)
}

type stateService struct {
	c           client.Client
	serviceName string
}

func StateServiceClient(serviceName string, c client.Client) StateService {
	if c == nil {
		c = client.NewClient()
	}
	if len(serviceName) == 0 {
		serviceName = "proto"
	}
	return &stateService{
		c:           c,
		serviceName: serviceName,
	}
}

func (c *stateService) Online(ctx context.Context, in *UserOnlineReq, opts ...client.CallOption) (*UserOnlineRes, error) {
	req := c.c.NewRequest(c.serviceName, "State.Online", in)
	out := new(UserOnlineRes)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stateService) Offline(ctx context.Context, in *UserOfflineReq, opts ...client.CallOption) (*UserOfflineRes, error) {
	req := c.c.NewRequest(c.serviceName, "State.Offline", in)
	out := new(UserOfflineRes)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for State service

type StateHandler interface {
	Online(context.Context, *UserOnlineReq, *UserOnlineRes) error
	Offline(context.Context, *UserOfflineReq, *UserOfflineRes) error
}

func RegisterStateHandler(s server.Server, hdlr StateHandler, opts ...server.HandlerOption) {
	s.Handle(s.NewHandler(&State{hdlr}, opts...))
}

type State struct {
	StateHandler
}

func (h *State) Online(ctx context.Context, in *UserOnlineReq, out *UserOnlineRes) error {
	return h.StateHandler.Online(ctx, in, out)
}

func (h *State) Offline(ctx context.Context, in *UserOfflineReq, out *UserOfflineRes) error {
	return h.StateHandler.Offline(ctx, in, out)
}
