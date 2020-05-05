// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	pb "do-tutorial/src/productservice/genproto"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	products pb.ListProductsResponse
	productMutex *sync.Mutex
	log *logrus.Logger
	extraLatency time.Duration

	port = "3050"

	reloadProducts bool

)

func init()  {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg: "message",
		},
		TimestampFormat: time.RFC3339Nano,
	}

	log.Out = os.Stdout
	productMutex = &sync.Mutex{}
	err := readProductFile(&products)
	if err != nil {
		log.Warnf("could not parse products ")
	}
}

func main()  {
	// set injected latency
	if s := os.Getenv("EXTRA_LATENCY"); s != "" {
		v, err := time.ParseDuration(s)
		if err != nil {
			log.Fatalf("failed to parse EXTRA_LATENCY (%s) as time.Duration: %+v", v, err)
		}

		extraLatency = v
		log.Infof("extra latency enabled (duration: %v", extraLatency)
	} else {
		extraLatency = time.Duration(0)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		for {
			sig := <- sigs
			log.Printf("received signal: %s", sig)
			if sig == syscall.SIGUSR1 {
				reloadProducts = true
				log.Infof("Enable products reloading")
			} else {
				reloadProducts = false
				log.Infof("Disable product reloading")
			}
		}
	}()

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	log.Infof("starting grpc server at :%s", port)

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	var srv *grpc.Server

	svc := &Products{}

	pb.RegisterProductServiceServer(srv, svc)
	go srv.Serve(l)
	fmt.Sprintf("%s",l.Addr().String())

	select {}
}

type Products struct {}



func (p Products) ListProducts(ctx context.Context, empty *pb.Empty) (*pb.ListProductsResponse, error) {
	panic("implement me")
}

func (p Products) GetProduct(ctx context.Context, request *pb.GetProductRequest) (*pb.Product, error) {
	panic("implement me")
}

func (p Products) SearchProducts(ctx context.Context, request *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	panic("implement me")
}


