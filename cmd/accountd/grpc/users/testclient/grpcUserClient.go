// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// This application is a simple gRPC accountd client. It runs a series of smoke tests against an accountd
// service accepting gRPC requests.
//
// It is intended to be run by executing ../../../../../smoketestStandAloneGRPC.sh.
//
// It can also be run manually if the following requirements are met:
//	1.	A running gRPC accountd service listening on 'localhost:5000' which 'may' be started using:
//		../../../accountd -configFile "testdata/config/config" -secretsDir "testdata/secrets" -protocol "grpc" &
//	2.	MySQL running on port 3306 or 6603 depending on the 'dbPort' value in ../../../testdata/config/config

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
	fmt.Printf("Users gRPC Client starting...\n\n")

	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:5000", opts)
	if err != nil {
		log.Fatalf("\terror %s attempting to connect to Accountd gRPC server", err)
	}
	defer cc.Close()

	client := pb.NewUserServerClient(cc)

	fmt.Println("Get all Users")
	fmt.Println("\tExpect 5 users")
	resp, _ := client.GetUsers(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Printf("\tGet all Users failed with error %s\n\n", err)
	}
	if resp == nil {
		fmt.Println("\tERROR: nil GetUsers response received")
	} else {
		fmt.Printf("\tGet all users response: %+v\n\n", resp.Users)
	}

	// Create new user to delete later
	fmt.Println("Create new user Brian Wilson")
	fmt.Printf("\tExpect new user Brian Wilson to be created\n")
	u := &pb.User{
		AccountID: 1,
		HREF:      "",
		Name:      "Brian Wilson",
		EMail:     "goodvibrations@gmail.com",
		Role:      pb.RoleEnum_UNRESTRICTED,
		Password:  "helpmerhonda",
	}
	id, err := client.CreateUser(context.Background(), u)
	if err != nil {
		fmt.Printf("\tCreateUser Brian Wilson with email %s failed with error %s\n", u.GetEMail(), err)
	} else {
		u.ID = id.GetId()
		fmt.Printf("\tCreate User Brian Wilson successful for user %d with email %s\n", id.GetId(), u.GetEMail())
	}

	fmt.Println("Get User")
	fmt.Println("\tExpect Brian Wilson")
	respBW, err := client.GetUser(context.Background(), id)
	if err != nil {
		fmt.Printf("\t Get User Brian Wilson with id %d failed with error %s\n", id.GetId(), err)
	}
	if respBW == nil {
		fmt.Println("\tnil GetUser Brian Wilson after CreateUser response received")
	} else {
		fmt.Printf("\tGet User Brian  after CreateUser response: %+v\n\n", respBW.String())
	}

	fmt.Println("Create duplicate user Brian Wilson")
	fmt.Printf("\tExpect duplicate user Brian Wilson will not be created\n")
	u2 := &pb.User{
		AccountID: 1,
		HREF:      "",
		Name:      "Brian Wilson",
		EMail:     "goodvibrations@gmail.com",
		Role:      pb.RoleEnum_UNRESTRICTED,
		Password:  "helpmerhonda",
	}
	id, err = client.CreateUser(context.Background(), u2)
	if err != nil {
		fmt.Printf("\tDuplicate Brian Wilson with email %s failed with error %s\n", u2.GetEMail(), err)
	} else {
		fmt.Printf("\tUnexpected sucess creating duplicate User Brian Wilson successful for user %d with email %s\n", id.GetId(), u2.GetEMail())
	}

	fmt.Println("Update User Brian Wilson")
	fmt.Println("\tExpect Beach Boy Brian Wilson")
	respBW.Name = "Beach Boy Brian Wilson"
	respBW.Password = "helpmerhonda"
	_, err = client.UpdateUser(context.Background(), respBW)
	if err != nil {
		fmt.Printf("\tUpdate User Brian Wilson with id %d failedwith error %s\n", respBW.GetID(), err)
	}
	if respBW == nil {
		fmt.Println("\tnil Update Beach Boy Brian Wilson after CreateUser response received")
	} else {
		fmt.Printf("\tGet User Beach Boy Brian Wilson  after UpdateUser response: %+v\n", respBW.String())
	}

	fmt.Println("Delete User Beach Boy Brian Wilson")
	fmt.Println("\tExpect Beach Boy Brian Wilson to be deleted")
	_, err = client.DeleteUser(context.Background(), &pb.UserID{Id: respBW.ID})
	if err != nil {
		fmt.Printf("\tDeleteUser Beach Boy Brian Wilson %d with failed\n", respBW.ID)
	} else {
		fmt.Printf("\tDeleteUser Beach Boy Brian Wilson successful for user %d \n", respBW.ID)
	}

	fmt.Println("Get all Users")
	fmt.Println("\tExpect Beach Boy Brian Wilson to be deleted")
	resp, err = client.GetUsers(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Printf("\tGet all Users failed with error %s\n\n", err)
	}
	if resp == nil {
		fmt.Println("\tERROR: nil GetUsers response received")
	} else {
		fmt.Printf("\tGet all users response: %+v\n\n", resp.Users)
	}

	fmt.Println("Get User")
	fmt.Println("\tExpect User not found error")
	resp2, err := client.GetUser(context.Background(), &pb.UserID{Id: int64(105)})
	if err != nil {
		fmt.Printf("\tExpected nil GetUser response received, error %s\n\n", err)
	} else {
		fmt.Printf("\tUnexpected GetUser success response: %+v\n\n", resp2)
	}

	fmt.Println("Update non-existing User Brian Wilson")
	fmt.Println("\tExpect error")
	respBW.Password = "password"
	_, err = client.UpdateUser(context.Background(), respBW)
	if err != nil {
		fmt.Printf("\tExpected update of non-existent User Brian Wilson as expected with error %s\n\n", err)
	} else {
		fmt.Printf("\tUnexpected successful update of non-existent User Brian Wilson\n\n")
	}

}
