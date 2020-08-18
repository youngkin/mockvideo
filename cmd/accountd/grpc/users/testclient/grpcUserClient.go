// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/youngkin/mockvideo/pkg/accountd"
)

func main() {
	fmt.Println("Users gRPC Client starting...")

	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:5000", opts)
	if err != nil {
		log.Fatalf("error %s attempting to connect to Accountd gRPC server", err)
	}
	defer cc.Close()

	client := pb.NewUserClient(cc)

	resp, _ := client.GetUsers(context.Background(), &empty.Empty{})
	fmt.Printf("GetUsers response: %+v\n\n", resp.Users)

	resp2, _ := client.GetUser(context.Background(), &pb.GetUserRqst{Id: int64(1)})
	fmt.Printf("GetUser response: %+v\n\n", resp2)
}
