package payment

import (
	"context"

	"github.com/irfan-ghzl/pintour/internal/domain/entity"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type Gateway interface {
	CreateTransaction(ctx context.Context, req entity.PaymentRequest) (*entity.PaymentResult, error)
}

type midtransGateway struct {
	serverKey    string
	isProduction bool
}

func NewMidtransGateway(serverKey string, isProduction bool) Gateway {
	return &midtransGateway{
		serverKey:    serverKey,
		isProduction: isProduction,
	}
}

func (g *midtransGateway) CreateTransaction(ctx context.Context, req entity.PaymentRequest) (*entity.PaymentResult, error) {
	env := midtrans.Sandbox
	if g.isProduction {
		env = midtrans.Production
	}

	snapClient := snap.Client{}
	snapClient.New(g.serverKey, env)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.BookingCode,
			GrossAmt: int64(req.Amount),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: req.UserName,
			Email: req.UserEmail,
			Phone: req.UserPhone,
		},
	}

	snapResp, err := snapClient.CreateTransaction(snapReq)
	if err != nil {
		return nil, err
	}

	return &entity.PaymentResult{
		Token: snapResp.Token,
		URL:   snapResp.RedirectURL,
	}, nil
}
