package summerquickapi

import (
	"encoding/json"
	"fmt"

	"github.com/Meduzz/quickapi/model"
	"github.com/Meduzz/rpc"
	sqapi "github.com/Meduzz/summer-quickapi/api"
	"github.com/Meduzz/summer-quickapi/internal/proxy"
	"github.com/Meduzz/summer-quickapi/qa"
	summerrpc "github.com/Meduzz/summer-rpc"
	"github.com/Meduzz/summer/api"
	"github.com/Meduzz/summer/errors"
	"github.com/Meduzz/summer/framework"
	"gorm.io/gorm"
)

type (
	Wrapper struct {
		instance *framework.Summer
	}
)

// post api -> jsonrpc -> (method, params) -> (verb, params)
func QuickapiHttpProxy(base, entity string) {
	wrap := Wrap(framework.Instance)
	wrap.HttpProxy(base, entity)
}

// (post rpc) -> jsonrpc -> (method, params) -> (topic, params)
func QuickapiRPCProxy(client *rpc.RPC, prefix, entity string, timeout int) {
	wrap := Wrap(framework.Instance)
	wrap.RpcProxy(client, prefix, entity, timeout)
}

// (post rpc) -> jsonrpc -> (method, params) -> storer
func Quickapi(db *gorm.DB, entity model.Entity) {
	wrap := Wrap(framework.Instance)
	wrap.Quickapi(db, entity)
}

func Wrap(instance *framework.Summer) *Wrapper {
	return &Wrapper{
		instance: instance,
	}
}

func (w *Wrapper) HttpProxy(base, entity string) {
	client := qa.NewHttpClient(base, entity)
	setup(w.instance, entity, client)
}

func (w *Wrapper) RpcProxy(rpcClient *rpc.RPC, prefix, entity string, timeout int) {
	setupRpc(w.instance, rpcClient, prefix, entity, timeout)
}

func (w *Wrapper) Quickapi(db *gorm.DB, entity model.Entity) {
	storer, _ := proxy.NewLocalProxy(db, entity) // TODO ignored error is ignored...
	setup(w.instance, entity.Name(), storer)
}

func setup(instance *framework.Summer, entity string, proxy sqapi.Proxy) {
	register(instance, method(entity, "create"), genericReqRes(proxy.Create))
	register(instance, method(entity, "update"), genericReqRes(proxy.Update))
	register(instance, method(entity, "read"), genericReqRes(proxy.Read))
	register(instance, method(entity, "delete"), genericReq(proxy.Delete))
	register(instance, method(entity, "search"), genericReqRes(proxy.Search))
	register(instance, method(entity, "patch"), genericReqRes(proxy.Patch))
}

func setupRpc(instance *framework.Summer, conn *rpc.RPC, prefix, entity string, timeout int) {
	register(instance, method(entity, "create"), summerrpc.RpcProxy(conn, topic(prefix, entity, "create"), timeout))
	register(instance, method(entity, "update"), summerrpc.RpcProxy(conn, topic(prefix, entity, "update"), timeout))
	register(instance, method(entity, "read"), summerrpc.RpcProxy(conn, topic(prefix, entity, "read"), timeout))
	register(instance, method(entity, "delete"), summerrpc.RpcProxy(conn, topic(prefix, entity, "delete"), timeout))
	register(instance, method(entity, "search"), summerrpc.RpcProxy(conn, topic(prefix, entity, "search"), timeout))
	register(instance, method(entity, "patch"), summerrpc.RpcProxy(conn, topic(prefix, entity, "patch"), timeout))
}

func genericReqRes[T any](handler func(*T) (any, error)) api.Handler {
	return func(r *api.Request) *api.Response {
		req := new(T)
		err := json.Unmarshal(r.Params, req)

		if err != nil {
			return framework.ErrorResponse(r.ID, errors.ParseError(err))
		}

		res, err := handler(req)

		if err != nil {
			return framework.ErrorResponse(r.ID, errors.ParseError(err))
		}

		return framework.ResultResponse(r.ID, res)
	}
}

func genericReq[T any](handler func(*T) error) api.Handler {
	return func(r *api.Request) *api.Response {
		req := new(T)
		err := json.Unmarshal(r.Params, req)

		if err != nil {
			return framework.ErrorResponse(r.ID, errors.ParseError(err))
		}

		err = handler(req)

		if err != nil {
			return framework.ErrorResponse(r.ID, errors.ParseError(err))
		}

		return &api.Response{
			ID:      r.ID,
			JsonRPC: r.JsonRPC,
			Result:  json.RawMessage([]byte("{}")), // TODO this... 🤬
		}
	}
}

func register(instance *framework.Summer, method string, handler api.Handler) {
	instance.Register(method, handler)
}

func method(entity, action string) string {
	return fmt.Sprintf("%s.%s", entity, action)
}

func topic(prefix, entity, action string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, entity, action)
}
