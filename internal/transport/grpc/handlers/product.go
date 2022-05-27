package grpcHandler

import (
	"context"
	"math"
	"net/http"

	"github.com/ernur-eskermes/product-store/internal/core"
	pb "github.com/ernur-eskermes/product-store/pkg/domain"
	"github.com/ernur-eskermes/product-store/pkg/filters"
	"github.com/gocarina/gocsv"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Product struct {
	service ProductService
	pb.UnimplementedProductServiceServer
}

func NewProduct(service ProductService) *Product {
	return &Product{
		service: service,
	}
}

func (h *Product) Fetch(ctx context.Context, req *pb.FetchRequest) (*empty.Empty, error) {
	resp, err := http.Get(req.Url)
	if err != nil {
		return &empty.Empty{}, status.Error(codes.NotFound, err.Error())
	}
	defer resp.Body.Close()

	products := make([]core.Product, 0)
	if err = gocsv.Unmarshal(resp.Body, &products); err != nil {
		return &empty.Empty{}, status.Error(codes.InvalidArgument, err.Error())
	}

	if err = h.service.UpdateOrCreate(ctx, products); err != nil {
		return &empty.Empty{}, status.Error(codes.Unknown, err.Error())
	}

	return &empty.Empty{}, nil
}

func (h *Product) List(ctx context.Context, req *pb.Filters) (*pb.ListResponse, error) {
	f := filters.New(
		req.Page,
		req.PageSize,
		req.Sort,
		"name",
		[]string{"name", "-name", "price", "-price"},
	)
	if err := filters.ValidateFilters(f); err != nil {
		return nil, ErrorFilterResponse(err)
	}

	products, err := h.service.GetAll(ctx, f)
	if err != nil {
		return nil, err
	}

	totalRecords, err := h.service.GetTotalRecords(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]*pb.ListResponse_Product, 0, len(products))

	for _, product := range products {
		res = append(res, &pb.ListResponse_Product{Name: product.Name, Price: int64(product.Price)})
	}

	return &pb.ListResponse{
		Results:  res,
		Metadata: calculateMetadata(totalRecords, f.Page, f.PageSize),
	}, nil
}

func calculateMetadata(totalRecords, page, pageSize int64) *pb.ListResponse_MetaData {
	if totalRecords == 0 {
		return nil
	}

	return &pb.ListResponse_MetaData{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int64(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
