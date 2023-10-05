/*
 * Copyright 2023 github.com/fatima-go
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @project fatima-core
 * @author jin
 * @date 23. 4. 14. 오후 6:07
 */

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatima-go/fatima-core"
	"github.com/fatima-go/fatima-core/builder"
	proto "github.com/fatima-go/fatima-core/builder/fatima.message.v1"
	"github.com/fatima-go/fatima-log"
	"github.com/fatima-go/saturn/domain"
	"github.com/fatima-go/saturn/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"net"
	"sync/atomic"
	"time"
)

const (
	propPredefineSaturnPort = "var.saturn.port"
	valueDefaultAddress     = ":4389"
)

func NewGrpcServer(fatimaRuntime fatima.FatimaRuntime, applicationExecutor service.ApplicationExecutor) *GrpcServer {
	server := new(GrpcServer)
	server.fatimaRuntime = fatimaRuntime
	server.applicationExecutor = applicationExecutor
	return server
}

type GrpcServer struct {
	fatimaRuntime       fatima.FatimaRuntime
	applicationExecutor service.ApplicationExecutor
	listener            net.Listener
	server              *grpc.Server
	proto.UnimplementedFatimaMessageServiceServer
}

func (g *GrpcServer) Initialize() bool {
	log.Info("GrpcServer Initialize()")

	address := g.buildServiceAddress()

	var err error
	g.listener, err = net.Listen("tcp", address)
	if err != nil {
		log.Warn("failed to listen: %v", err)
		return false
	}

	log.Info("grpc.controller server start. address=%v", address)

	// create server
	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.UnaryInterceptor(unaryInterceptor)}
	g.server = grpc.NewServer(opts...)

	// regist controllers
	proto.RegisterFatimaMessageServiceServer(g.server, g)

	reflection.Register(g.server)

	go func() {
		err = g.server.Serve(g.listener)
	}()

	time.Sleep(time.Millisecond * 200)
	if err != nil {
		log.Warn("serving error : %s", err.Error())
		return false
	}

	return true
}

func (g *GrpcServer) Bootup() {
	log.Info("GrpcServer Bootup()")
}

func (g *GrpcServer) Goaway() {
}

func (g *GrpcServer) Shutdown() {
	log.Info("GrpcServer Shutdown()")
	g.server.Stop()
}

func (g *GrpcServer) GetType() fatima.FatimaComponentType {
	return fatima.COMP_READER
}

func (g *GrpcServer) buildServiceAddress() string {
	fproc, ok := g.fatimaRuntime.(*builder.FatimaRuntimeProcess)
	if !ok {
		return valueDefaultAddress
	}

	address, ok := fproc.GetBuilder().GetPredefines().GetDefine(propPredefineSaturnPort)
	if !ok {
		return valueDefaultAddress
	}

	return address
}

func (g *GrpcServer) SendFatimaMessage(ctx context.Context, request *proto.SendFatimaMessageRequest) (*proto.SendFatimaMessageResponse, error) {
	response := &proto.SendFatimaMessageResponse{}
	var err error
	g.consume(request.GetJsonString())
	if err != nil {
		response.Response = &proto.SendFatimaMessageResponse_Error{Error: buildInternalServerError(err)}

		log.Warn("SendFatimaMessage error : %s", err.Error())
		return response, nil
	}

	response.Response = &proto.SendFatimaMessageResponse_Success{Success: &proto.ResponseSuccess{}}
	return response, nil

}

func (g *GrpcServer) consume(jsonString string) {
	if len(jsonString) < 3 {
		return
	}

	if log.IsTraceEnabled() {
		log.Trace("%s", jsonString)
	}

	data := []byte(jsonString)
	var message domain.MBusMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		log.Warn("fail to unmarshal : %s", err.Error())
		return
	}

	if message.Header.ApplicationCode != builder.ApplicationCode {
		return
	}

	g.applicationExecutor.Consume(message)
}

func buildInternalServerError(err error) *proto.ResponseError {
	return &proto.ResponseError{
		GrpcResponse: proto.ResponseError_SERVER_ERROR,
		Code:         proto.ResponseError_ERROR_ETC,
		Value:        fmt.Sprintf("Internal Server Error : %s", err.Error())}
}

// unaryInterceptor calls authenticateClient with current context
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var remoteAddress string
	p, ok := peer.FromContext(ctx)
	if ok {
		remoteAddress = p.Addr.String()
		ctx = context.WithValue(ctx, "remote-addr", p.Addr.String())
	}
	transactionId := fmt.Sprintf("CSTM%06d", nextCustomTransactionId())
	ctx = context.WithValue(ctx, "transactionId", transactionId)
	log.Debug("<--- [%12s] --- : [%s]", transactionId, remoteAddress)
	return handler(ctx, req)
}

var ops uint64

func nextCustomTransactionId() uint64 {
	next := atomic.AddUint64(&ops, 1)
	if next > 900000 {
		atomic.StoreUint64(&ops, 0)
	}
	return next
}
