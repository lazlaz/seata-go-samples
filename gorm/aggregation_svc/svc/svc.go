package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

import (
	"github.com/opentrx/seata-go-samples/order_svc/dao"
	dao2 "github.com/opentrx/seata-go-samples/product_svc/dao"
	context2 "github.com/transaction-wg/seata-golang/pkg/client/context"
	"github.com/transaction-wg/seata-golang/pkg/client/tm"
)

type Svc struct {
}

func (svc *Svc) CreateSo(ctx context.Context, rollback bool) error {
	rootContext := ctx.(*context2.RootContext)
	soMasters := []*dao.SoMaster{
		{
			BuyerUserSysno:       10001,
			SellerCompanyCode:    "SC001",
			ReceiveDivisionSysno: 110105,
			ReceiveAddress:       "朝阳区长安街001号",
			ReceiveZip:           "000001",
			ReceiveContact:       "斯密达",
			ReceiveContactPhone:  "18728828296",
			StockSysno:           1,
			PaymentType:          1,
			SoAmt:                430.5,
			Status:               10,
			Appid:                "dk-order",
			SoItems: []*dao.SoItem{
				{
					ProductSysno:  1,
					ProductName:   "刺力王",
					CostPrice:     200,
					OriginalPrice: 232,
					DealPrice:     215.25,
					Quantity:      2,
				},
			},
		},
	}

	reqs := []*dao2.AllocateInventoryReq{{
		ProductSysNo: 1,
		Qty:          2,
	}}

	type rq1 struct {
		Req []*dao.SoMaster
	}

	type rq2 struct {
		Req []*dao2.AllocateInventoryReq
	}

	q1 := &rq1{Req: soMasters}
	soReq, err := json.Marshal(q1)
	fmt.Println(string(soReq))
	req1, err := http.NewRequest("POST", "http://localhost:8002/createSo", bytes.NewBuffer(soReq))
	if err != nil {
		panic(err)
	}
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("xid", rootContext.GetXID())

	client := &http.Client{}
	result1, err1 := client.Do(req1)
	if err1 != nil {
		return err1
	}

	if result1.StatusCode == 400 {
		return errors.New("err")
	}

	q2 := &rq2{
		Req: reqs,
	}
	ivtReq, _ := json.Marshal(q2)
	fmt.Println(string(ivtReq))
	req2, err := http.NewRequest("POST", "http://localhost:8001/allocateInventory", bytes.NewBuffer(ivtReq))
	if err != nil {
		panic(err)
	}
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("xid", rootContext.GetXID())

	result2, err2 := client.Do(req2)
	if err2 != nil {
		return err2
	}

	if result2.StatusCode == 400 {
		return errors.New("err")
	}

	if rollback {
		return errors.New("there is a error")
	}
	return nil
}

var service = &Svc{}

type ProxyService struct {
	*Svc
	CreateSo func(ctx context.Context, rollback bool) error
}

var methodTransactionInfo = make(map[string]*tm.TransactionInfo)

func init() {
	methodTransactionInfo["CreateSo"] = &tm.TransactionInfo{
		TimeOut:     60000000,
		Name:        "CreateSo",
		Propagation: tm.REQUIRED,
	}
}

func (svc *ProxyService) GetProxyService() interface{} {
	return svc.Svc
}

func (svc *ProxyService) GetMethodTransactionInfo(methodName string) *tm.TransactionInfo {
	return methodTransactionInfo[methodName]
}

var ProxySvc = &ProxyService{
	Svc: service,
}
