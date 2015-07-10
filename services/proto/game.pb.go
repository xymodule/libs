// Code generated by protoc-gen-go.
// source: game.proto
// DO NOT EDIT!

package proto

import proto1 "github.com/golang/protobuf/proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal

type Game_FrameType int32

const (
	Game_Message    Game_FrameType = 0
	Game_Register   Game_FrameType = 1
	Game_Unregister Game_FrameType = 2
	Game_Kick       Game_FrameType = 3
)

var Game_FrameType_name = map[int32]string{
	0: "Message",
	1: "Register",
	2: "Unregister",
	3: "Kick",
}
var Game_FrameType_value = map[string]int32{
	"Message":    0,
	"Register":   1,
	"Unregister": 2,
	"Kick":       3,
}

func (x Game_FrameType) String() string {
	return proto1.EnumName(Game_FrameType_name, int32(x))
}

type Game struct {
}

func (m *Game) Reset()         { *m = Game{} }
func (m *Game) String() string { return proto1.CompactTextString(m) }
func (*Game) ProtoMessage()    {}

type Game_Frame struct {
	Type    Game_FrameType `protobuf:"varint,1,opt,enum=proto.Game_FrameType" json:"Type,omitempty"`
	Content []byte         `protobuf:"bytes,2,opt,proto3" json:"Content,omitempty"`
	UserId  int32          `protobuf:"varint,3,opt" json:"UserId,omitempty"`
}

func (m *Game_Frame) Reset()         { *m = Game_Frame{} }
func (m *Game_Frame) String() string { return proto1.CompactTextString(m) }
func (*Game_Frame) ProtoMessage()    {}

func init() {
	proto1.RegisterEnum("proto.Game_FrameType", Game_FrameType_name, Game_FrameType_value)
}

// Client API for GameService service

type GameServiceClient interface {
	Stream(ctx context.Context, opts ...grpc.CallOption) (GameService_StreamClient, error)
}

type gameServiceClient struct {
	cc *grpc.ClientConn
}

func NewGameServiceClient(cc *grpc.ClientConn) GameServiceClient {
	return &gameServiceClient{cc}
}

func (c *gameServiceClient) Stream(ctx context.Context, opts ...grpc.CallOption) (GameService_StreamClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_GameService_serviceDesc.Streams[0], c.cc, "/proto.GameService/Stream", opts...)
	if err != nil {
		return nil, err
	}
	x := &gameServiceStreamClient{stream}
	return x, nil
}

type GameService_StreamClient interface {
	Send(*Game_Frame) error
	Recv() (*Game_Frame, error)
	grpc.ClientStream
}

type gameServiceStreamClient struct {
	grpc.ClientStream
}

func (x *gameServiceStreamClient) Send(m *Game_Frame) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gameServiceStreamClient) Recv() (*Game_Frame, error) {
	m := new(Game_Frame)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for GameService service

type GameServiceServer interface {
	Stream(GameService_StreamServer) error
}

func RegisterGameServiceServer(s *grpc.Server, srv GameServiceServer) {
	s.RegisterService(&_GameService_serviceDesc, srv)
}

func _GameService_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GameServiceServer).Stream(&gameServiceStreamServer{stream})
}

type GameService_StreamServer interface {
	Send(*Game_Frame) error
	Recv() (*Game_Frame, error)
	grpc.ServerStream
}

type gameServiceStreamServer struct {
	grpc.ServerStream
}

func (x *gameServiceStreamServer) Send(m *Game_Frame) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gameServiceStreamServer) Recv() (*Game_Frame, error) {
	m := new(Game_Frame)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _GameService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.GameService",
	HandlerType: (*GameServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _GameService_Stream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
}
