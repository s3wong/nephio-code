/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
    "context"
    "fmt"
    "net"

    pb "github.com/nephio-experimental/nephio-sdk-poc/nfdeploymentrpc"
    "github.com/nephio-experimental/nephio-sdk-poc/free5gc-grpc-server/upf/vendor"
    nephiov1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"

    "google.golang.org/grpc"
    "k8s.io/client-go/rest"
    "sigs.k8s.io/controller-runtime/pkg/client"
	//"sigs.k8s.io/controller-runtime/pkg/client/config"
    //"sigs.k8s.io/controller-runtime/pkg/log"
)

type nfDeploymentServer struct {
    pb.UnimplementedNFDeploymentRPCServer
    client.Client
}

func newServer() *nfDeploymentServer {
    config, err := rest.InClusterConfig()
    if err != nil {
        fmt.Printf("ERROR: failed to create incluster config: %v\n", err)
        return nil
    }

    cl, err := client.New(config, client.Options{})
    if err != nil {
        //log.Error(err, "failed to create client")
        fmt.Printf("ERROR: failed to create client: %v\n", err)
        return nil
    }

    s := &nfDeploymentServer{
        Client: cl,
    }
    return s
}

func (s *nfDeploymentServer) GetNFDeployment(m *pb.NFDeployment) (*nephiov1alpha1.NFDeployment, error) {
    nfDeployment := new(nephiov1alpha1.NFDeployment)

    err := s.Client.Get(context.TODO(), client.ObjectKey{
        Namespace: m.Namespace,
        Name: m.Name,}, nfDeployment)
    if err != nil {
        //log.Error(err, "failed to get NFDeployment")
        fmt.Printf("ERROR: failed to get NFDeployment: %v\n", err)
        return nil, err
    }
    return nfDeployment, nil
}

func (s *nfDeploymentServer) CreateUpdate(ctx context.Context, m *pb.NFDeployment) (*pb.NFDeploymentResponse, error) {
    nfDeployment, err := s.GetNFDeployment(m)
    if err != nil {
        return nil, err
    }
    rsp, err := vendor.CreateUpdate(ctx, s.Client, nfDeployment)
    if err != nil {
        //log.Error(err, "CreateUpdate failed")
        fmt.Printf("ERROR: CreateUpdate failed: %v\n", err)
        return nil, err
    }

    return rsp, nil
}

func (s *nfDeploymentServer) Delete(ctx context.Context, m *pb.NFDeployment) (*pb.NFDeploymentResponse, error) {
    nfDeployment, err := s.GetNFDeployment(m)
    if err != nil {
        return nil, err
    }
    rsp, err := vendor.Delete(ctx, s.Client, nfDeployment)
    if err != nil {
        //log.Error(err, "Delete failed")
        fmt.Printf("ERROR: Delete failed: %v\n", err)
        return nil, err
    }

    return rsp, nil
}

func main() {

    s := newServer()
    if s == nil {
        //log.Fatal("Unable to create new server")
        fmt.Printf("FATAL: Unable to create new server")
        panic("Failed to create a new gRPC server")
    }
    // TODO(s3wong): get server port number from K8s API server
    portNum := "50051"
    lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", portNum))
    if err != nil {
        //log.Fatalf("failed to listen: %v", err)
        fmt.Printf("FATAL: failed to listen: %v", err)
        panic(err)
    }
    grpcServer := grpc.NewServer()
    pb.RegisterNFDeploymentRPCServer(grpcServer, s)
    grpcServer.Serve(lis)
}
