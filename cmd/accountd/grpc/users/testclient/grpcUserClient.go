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
	"github.com/youngkin/mockvideo/internal/domain"
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

	client := pb.NewUserServerClient(cc)

	resp, _ := client.GetUsers(context.Background(), &empty.Empty{})
	if resp == nil {
		fmt.Println("nil GetUsers response received")
	} else {
		fmt.Printf("GetUsers response: %+v\n\n", resp.Users)
	}

	resp2, _ := client.GetUser(context.Background(), &pb.UserID{Id: int64(2)})
	if resp2 == nil {
		fmt.Println("nil GetUser response received")
	} else {
		fmt.Printf("GetUser response: %+v\n\n", resp2)
	}

	resp2.Name = "peter tork 2"
	resp2.Password = "stupidpassword"
	resp2.Role = pb.RoleType(domain.Unrestricted)
	_, err = client.UpdateUser(context.Background(), resp2)
	if err != nil {
		fmt.Printf("UpdateUser %d with email %s failed with error %s\n", resp2.GetID(), resp2.GetEMail(), err)
	} else {
		fmt.Printf("UpdateUser successful for user %d with email %s\n", resp2.GetID(), resp2.GetEMail())
	}

	resp3, err := client.GetUser(context.Background(), &pb.UserID{Id: resp2.ID})
	if resp3 == nil {
		fmt.Printf("UpdateUser, nil GetUser response received with error %s\n", err)
	} else {
		fmt.Printf("UpdateUser response: %+v\n\n", resp3)
	}

	// Revert Update
	resp2.Name = "peter tork"
	resp2.Password = "alksdf98423)*(&#"
	_, err = client.UpdateUser(context.Background(), resp2)
	if err != nil {
		fmt.Printf("UpdateUser %d with email %s failed with error %s\n", resp2.GetID(), resp2.GetEMail(), err)
	} else {
		fmt.Printf("UpdateUser successful for user %d with email %s\n", resp2.GetID(), resp2.GetEMail())
	}

	// Create new user to delete later
	u := &pb.User{
		AccountID: 1,
		HREF:      "",
		Name:      "porgy tirebiter",
		EMail:     "pTirebiter@gmail.com",
		Role:      pb.RoleType_PRIMARY,
		Password:  "nickdangerthirdeye",
	}
	id, err := client.CreateUser(context.Background(), u)
	if err != nil {
		fmt.Printf("CreateUser with email %s failed\n", u.GetEMail())
	} else {
		u.ID = id.GetId()
		fmt.Printf("CreateUser successful for user %d with email %s\n", id.GetId(), u.GetEMail())
	}

	resp, _ = client.GetUsers(context.Background(), &empty.Empty{})
	if resp == nil {
		fmt.Println("nil GetUsers after CreateUser response received")
	} else {
		fmt.Printf("GetUsers after CreateUser response: %+v\n\n", resp.Users)
	}

	_, err = client.DeleteUser(context.Background(), id)
	if err != nil {
		fmt.Printf("DeleteUser %d with failed\n", id.GetId())
	} else {
		fmt.Printf("DeleteUser successful for user %d \n", id.GetId())
	}

	resp, _ = client.GetUsers(context.Background(), &empty.Empty{})
	if resp == nil {
		fmt.Println("nil GetUsers after reverting CreateUser response received")
	} else {
		fmt.Printf("GetUsers after reverting CreateUser response: %+v\n\n", resp.Users)
	}

}
