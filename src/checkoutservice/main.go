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
//
// Modifications made:
// Modify attached service addresses to suite project usecase

package checkoutservice

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"

	pb "do-tutorial/src/checkoutservice/genproto"
)

const (
	listenPort = "5050"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339,
	}
}

type checkoutService struct {
	productSvcAddr string
	cartSvcAddr    string
	userSvcAddr    string
}

func main() {
	port := listenPort
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	svc := new(checkoutService)
	mustMapEnv(&svc.productSvcAddr, "PRODUCT_SERVICE_ADDR")
	mustMapEnv(&svc.cartSvcAddr, "CART_SERVICE_ADDR")
	mustMapEnv(&svc.userSvcAddr, "USER_SERVICE_ADDR")

	log.Infof("service config: %+v", svc)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s", port))
	if err != nil {
		log.Fatal(err)
	}

	var srv *grpc.Server
	if os.Getenv("DISABLE_STATS") == "" {
		log.Info("stats enabled.")
		srv = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	} else {
		log.Info("stats disabled")
		srv = grpc.NewServer()
	}

	pb.RegisterCheckoutServiceServer(srv, svc)
	log.Infof("starting to listen on tcp: %q", lis.Addr().String())
	err = srv.Serve(lis)
	log.Fatal(err)

}

func mustMapEnv(target *string, envKey string) {
	v := os.Getenv(envKey)
	if v == "" {
		panic(fmt.Sprintf("environment variable %q not set", envKey))
	}
	*target = v
}

func (c *checkoutService) PlaceOrder(ctx context.Context, request *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	log.Info("[PlaceOrder] user_id=%q", req.userId)

	orderId, err := uuid.NewUUID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate order uuid")
	}

	//prepareOrderItemsAndShippingQuoteFromCart
	prep, err := c.prepareOrderItemsFromCart(ctx, request.UserId, request.User.Address)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	total := pb.Money{
		CurrencyCode: "USD",
		Units: 0,
		Nanos: 0,
	}

	for _, it := range prep.orderItems {
		//total =
	}

	orderResult := &pb.OrderResult{
		OrderId: orderId.String(),
		Items: prep.orderItems,
		ShippingAddress: request.User.Address,
	}

	resp := &pb.PlaceOrderResponse{Order: orderResult}

	return resp, nil
}

type orderPrep struct {
	orderItems []*pb.OrderItem
	cartItems  []*pb.CartItem
}

func (c *checkoutService) prepareOrderItemsFromCart(ctx context.Context, userId string, address *pb.Address) (orderPrep, error) {
	var out orderPrep
	cartItems, err := c.getUserCart(ctx, userId)
	if err != nil {
		return out, fmt.Errorf("cart failure: %+v", err)
	}

	orderItems, err := c.prepOrderItems(ctx, cartItems)
	if err != nil {
		return out, fmt.Errorf("failed to prepare order: %+v", err)
	}

	out.cartItems = cartItems
	out.orderItems = orderItems
	return out, nil
}

func (c *checkoutService) getUserCart(ctx context.Context, userId string) ([]*pb.CartItem, error) {
	conn, err := grpc.DialContext(ctx, c.cartSvcAddr, grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{}))
	if err != nil {
		return nil, fmt.Errorf("could not connect cart service: %+v", err)
	}
	defer conn.Close()

	cart, err := pb.NewCartServiceClient(conn).GetCart(ctx, &pb.GetCartRequest{UserId: userId})
	if err != nil {
		return nil, fmt.Errorf("failed to get user cart during checkout: %+v", err)
	}

	return cart.GetItems(), nil
}

func (c *checkoutService) emptyUserCart(ctx context.Context, userId string) error {
	conn, err := grpc.DialContext(ctx, c.cartSvcAddr, grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{}))
	if err != nil {
		return fmt.Errorf("could not connect cart service: %+v", err)
	}
	defer conn.Close()

	if _, err := pb.NewCartServiceClient(conn).EmptyCart(ctx, &pb.EmptyCartRequest{UserId: userId}); err != nil {
		return fmt.Errorf("failed to empty user cart during checkout %+v", err)
	}

	return nil
}

func (c *checkoutService) prepOrderItems(ctx context.Context, items []*pb.CartItem) ([]*pb.OrderItem, error) {
	out := make([]*pb.OrderItem, len(items))

	conn, err := grpc.DialContext(ctx, c.productSvcAddr, grpc.WithInsecure(), grpc.WithStatsHandler(&ocgrpc.ClientHandler{}))
	if err != nil {
		return nil, fmt.Errorf("could not connect product service: %+v", err)
	}
	defer conn.Close()
	cl := pb.NewProductServiceClient(conn)

	for i, item := range items {
		product, err := cl.GetProduct(ctx, &pb.GetProductRequest{Id: item.GetProductId()})
		if err != nil {
			return nil, fmt.Errorf("failed to get product #%q", item.GetProductId())
		}

		price := product.GetPriceUsd()

		out[i] = &pb.OrderItem{
			Item: item,
			Cost: price,
		}
	}
	return out, nil
}

