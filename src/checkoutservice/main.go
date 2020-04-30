package checkoutservice

import (
	"context"
	"fmt"
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
	productSvcAddr  string
	cartSvcAddr     string
	currencySvcAddr string
	shippingSvcAddr string
	emailSvcAddr    string
	paymentSvcAddr  string
}

func main() {
	port := listenPort
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	svc := new(checkoutService)

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

func (c checkoutService) PlaceOrder(ctx context.Context, request *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	panic("implement me")
}
