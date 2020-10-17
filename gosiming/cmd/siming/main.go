package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "shaunybear/gosiming/internal/simac"

	"github.com/urfave/cli/v2"
	grpc "google.golang.org/grpc"
)

func cliPrintMacDetails(c *cli.Context) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewSiMacClient(conn)
	printMacDetails(client)

	return err
}

func printMacDetails(client pb.SiMacClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.GetMacDetails(ctx)
	if err != nil {
		log.Fatalf("%v.GetMacDetails(_) = _, %v", client, err)

	}

	fmt.Print("Client: Send MAC Details request")
	deveui := &pb.DevEui{Deveui: "Hi"}
	stream.Send(deveui)

	for {
		details, err := stream.Recv()
		if err == io.EOF {
			fmt.Printf("Client: Recv io.EOF\n")
			break
		}
		if err != nil {
			log.Fatalf("%v.GetMacDetails(_) = _, %v", client, err)
		}
		fmt.Printf("Client: Received %v\n", details)
	}
}

func main() {
	app := &cli.App{
		Name:   "siming",
		Usage:  "Siming client",
		Action: cliPrintMacDetails,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
