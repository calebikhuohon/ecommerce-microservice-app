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
	"bytes"
	"context"
	pb "do-tutorial/src/productservice/genproto"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	products     pb.ListProductsResponse
	productMutex *sync.Mutex
	log          *logrus.Logger
	extraLatency time.Duration

	port = "3050"

	reloadProducts bool
)

func init() {
	log = logrus.New()
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
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

func main() {
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
			sig := <-sigs
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
	if os.Getenv("DISABLE_STATS") == "" {
		log.Info("Stats enabled.")
		srv = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	} else {
		log.Info("Stats disabled.")
		srv = grpc.NewServer()
	}
	svc := &Products{}

	pb.RegisterProductServiceServer(srv, svc)
	log.Infof("starting to listen on tcp: %q", l.Addr().String())
	err = srv.Serve(l)
	log.Fatal(err)

	select {}
}

type Products struct{}

func readProductFile(products *pb.ListProductsResponse) error {
	productMutex.Lock()
	defer productMutex.Unlock()

	productJSON, err := ioutil.ReadFile("products.json")
	if err != nil {
		log.Fatalf("failed to open product json file: %v", err)
		return err
	}

	if err := jsonpb.Unmarshal(bytes.NewReader(productJSON), products); err != nil {
		log.Warnf("failed to parse the product JSON: %v", err)
		return err
	}
	log.Info("successfully parsed product catalog json")
	return nil
}

func parseProducts() []*pb.Product {
	if reloadProducts || len(products.Products) == 0 {
		err := readProductFile(&products)
		if err != nil {
			return []*pb.Product{}
		}
	}
	return products.Products
}

func (p *Products) ListProducts(ctx context.Context, empty *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(extraLatency)
	return &pb.ListProductsResponse{Products: parseProducts()}, nil
}

func (p *Products) GetProduct(ctx context.Context, request *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(extraLatency)
	var found *pb.Product
	for i := 0; i < len(parseProducts()); i++ {
		if request.Id == parseProducts()[i].Id {
			found = parseProducts()[i]
		}
	}

	if found == nil {
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", request.Id)
	}
	return found, nil
}

func (p *Products) SearchProducts(ctx context.Context, request *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	var ps []*pb.Product

	for _, p := range parseProducts() {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(request.Query)) ||
			strings.Contains(strings.ToLower(p.Description), strings.ToLower(request.Query)) {
			ps = append(ps, p)
		}
	}

	return &pb.SearchProductsResponse{Results: ps}, nil
}
