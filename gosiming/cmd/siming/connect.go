/*
 *
 * Copyright 2015 gRPC authors.
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
 */

// Package cmd implements a client for SiMac service.
package main

import (
	"context"
	"time"

	grpc "google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

// ServerConnect  Returns a server connection
func ServerConnect() (conn *grpc.ClientConn, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Set up a connection to the server.
	return grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
}
