// Code generated by protoc-gen-go.
// source: chat.proto
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

type Chat struct {
}

func (m *Chat) Reset()         { *m = Chat{} }
func (m *Chat) String() string { return proto1.CompactTextString(m) }
func (*Chat) ProtoMessage()    {}

type Chat_Nil struct {
}

func (m *Chat_Nil) Reset()         { *m = Chat_Nil{} }
func (m *Chat_Nil) String() string { return proto1.CompactTextString(m) }
func (*Chat_Nil) ProtoMessage()    {}

type Chat_Message struct {
	Id   uint64 `protobuf:"varint,1,opt" json:"Id,omitempty"`
	Body []byte `protobuf:"bytes,2,opt,proto3" json:"Body,omitempty"`
}

func (m *Chat_Message) Reset()         { *m = Chat_Message{} }
func (m *Chat_Message) String() string { return proto1.CompactTextString(m) }
func (*Chat_Message) ProtoMessage()    {}

type Chat_Id struct {
	Id uint64 `protobuf:"varint,1,opt" json:"Id,omitempty"`
}

func (m *Chat_Id) Reset()         { *m = Chat_Id{} }
func (m *Chat_Id) String() string { return proto1.CompactTextString(m) }
func (*Chat_Id) ProtoMessage()    {}

func init() {
}

// Client API for ChatService service

type ChatServiceClient interface {
	Subscribe(ctx context.Context, in *Chat_Id, opts ...grpc.CallOption) (ChatService_SubscribeClient, error)
	Read(ctx context.Context, in *Chat_Id, opts ...grpc.CallOption) (ChatService_ReadClient, error)
	Send(ctx context.Context, in *Chat_Message, opts ...grpc.CallOption) (*Chat_Nil, error)
	Reg(ctx context.Context, in *Chat_Id, opts ...grpc.CallOption) (*Chat_Nil, error)
}

type chatServiceClient struct {
	cc *grpc.ClientConn
}

func NewChatServiceClient(cc *grpc.ClientConn) ChatServiceClient {
	return &chatServiceClient{cc}
}

func (c *chatServiceClient) Subscribe(ctx context.Context, in *Chat_Id, opts ...grpc.CallOption) (ChatService_SubscribeClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ChatService_serviceDesc.Streams[0], c.cc, "/proto.ChatService/Subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatServiceSubscribeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ChatService_SubscribeClient interface {
	Recv() (*Chat_Message, error)
	grpc.ClientStream
}

type chatServiceSubscribeClient struct {
	grpc.ClientStream
}

func (x *chatServiceSubscribeClient) Recv() (*Chat_Message, error) {
	m := new(Chat_Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *chatServiceClient) Read(ctx context.Context, in *Chat_Id, opts ...grpc.CallOption) (ChatService_ReadClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_ChatService_serviceDesc.Streams[1], c.cc, "/proto.ChatService/Read", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatServiceReadClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ChatService_ReadClient interface {
	Recv() (*Chat_Message, error)
	grpc.ClientStream
}

type chatServiceReadClient struct {
	grpc.ClientStream
}

func (x *chatServiceReadClient) Recv() (*Chat_Message, error) {
	m := new(Chat_Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *chatServiceClient) Send(ctx context.Context, in *Chat_Message, opts ...grpc.CallOption) (*Chat_Nil, error) {
	out := new(Chat_Nil)
	err := grpc.Invoke(ctx, "/proto.ChatService/Send", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatServiceClient) Reg(ctx context.Context, in *Chat_Id, opts ...grpc.CallOption) (*Chat_Nil, error) {
	out := new(Chat_Nil)
	err := grpc.Invoke(ctx, "/proto.ChatService/Reg", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ChatService service

type ChatServiceServer interface {
	Subscribe(*Chat_Id, ChatService_SubscribeServer) error
	Read(*Chat_Id, ChatService_ReadServer) error
	Send(context.Context, *Chat_Message) (*Chat_Nil, error)
	Reg(context.Context, *Chat_Id) (*Chat_Nil, error)
}

func RegisterChatServiceServer(s *grpc.Server, srv ChatServiceServer) {
	s.RegisterService(&_ChatService_serviceDesc, srv)
}

func _ChatService_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Chat_Id)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChatServiceServer).Subscribe(m, &chatServiceSubscribeServer{stream})
}

type ChatService_SubscribeServer interface {
	Send(*Chat_Message) error
	grpc.ServerStream
}

type chatServiceSubscribeServer struct {
	grpc.ServerStream
}

func (x *chatServiceSubscribeServer) Send(m *Chat_Message) error {
	return x.ServerStream.SendMsg(m)
}

func _ChatService_Read_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Chat_Id)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChatServiceServer).Read(m, &chatServiceReadServer{stream})
}

type ChatService_ReadServer interface {
	Send(*Chat_Message) error
	grpc.ServerStream
}

type chatServiceReadServer struct {
	grpc.ServerStream
}

func (x *chatServiceReadServer) Send(m *Chat_Message) error {
	return x.ServerStream.SendMsg(m)
}

func _ChatService_Send_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(Chat_Message)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ChatServiceServer).Send(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func _ChatService_Reg_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(Chat_Id)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(ChatServiceServer).Reg(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _ChatService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ChatService",
	HandlerType: (*ChatServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Send",
			Handler:    _ChatService_Send_Handler,
		},
		{
			MethodName: "Reg",
			Handler:    _ChatService_Reg_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Subscribe",
			Handler:       _ChatService_Subscribe_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Read",
			Handler:       _ChatService_Read_Handler,
			ServerStreams: true,
		},
	},
}