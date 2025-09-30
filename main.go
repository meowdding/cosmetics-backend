package main

import (
	"cosmetics/routes"
	"cosmetics/routes/utils"
	"fmt"
	"net/http"
)

func setDefaults(route *RequestRoute) {
	if route.Get == nil {
		route.Get = NotImplementedRequestHandler{}
	}
	if route.Post == nil {
		route.Post = NotImplementedRequestHandler{}
	}
	if route.Put == nil {
		route.Put = NotImplementedRequestHandler{}
	}
	if route.Delete == nil {
		route.Delete = NotImplementedRequestHandler{}
	}
	if route.Patch == nil {
		route.Patch = NotImplementedRequestHandler{}
	}
}

type RequestRoute struct {
	Get    AbstractRequestHandler
	Post   AbstractRequestHandler
	Put    AbstractRequestHandler
	Delete AbstractRequestHandler
	Patch  AbstractRequestHandler
}

type AbstractRequestHandler interface {
	handle(http.ResponseWriter, *http.Request)
}

type NotImplementedRequestHandler struct{}

func authenticated(handler func(utils.RouteContext, http.ResponseWriter, *http.Request)) AuthenticatedRequestHandler {
	return AuthenticatedRequestHandler{handler: handler}
}

func public(handler func(utils.RouteContext, http.ResponseWriter, *http.Request)) RequestHandler {
	return RequestHandler{handler: handler}
}

type RequestHandler struct {
	handler func(utils.RouteContext, http.ResponseWriter, *http.Request)
}

type AuthenticatedRequestHandler struct {
	handler func(utils.RouteContext, http.ResponseWriter, *http.Request)
}

var routeContext = utils.NewRouteContext()

func (not NotImplementedRequestHandler) handle(res http.ResponseWriter, _ *http.Request) {
	res.WriteHeader(http.StatusMethodNotAllowed)
}

func (normal RequestHandler) handle(res http.ResponseWriter, req *http.Request) {
	normal.handler(routeContext, res, req)
}

func (authenticated AuthenticatedRequestHandler) handle(res http.ResponseWriter, req *http.Request) {
	if !utils.IsAuthenticated(req.Header.Get("Authorization")) {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}
	authenticated.handler(routeContext, res, req)
}

func create(handlers RequestRoute) func(http.ResponseWriter, *http.Request) {
	return createSave("", handlers)
}

func createSave(path string, handlers RequestRoute) func(http.ResponseWriter, *http.Request) {
	setDefaults(&handlers)
	return func(res http.ResponseWriter, req *http.Request) {
		if len(path) != 0 && req.URL.Path != path {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		switch req.Method {
		case "GET":
			handlers.Get.handle(res, req)
		case "POST":
			handlers.Post.handle(res, req)
		case "PUT":
			handlers.Put.handle(res, req)
		case "PATCH":
			handlers.Patch.handle(res, req)
		case "DELETE":
			handlers.Delete.handle(res, req)
		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	http.HandleFunc("/cosmetics", create(RequestRoute{
		Post:   authenticated(routes.CreateCosmetic),
		Put:    authenticated(routes.UpdateCosmetic),
		Delete: authenticated(routes.DeleteCosmetic),
	}))
	http.HandleFunc("/cosmetics/ids", create(RequestRoute{
		Get: public(routes.ListCosmeticIds),
	}))
	http.HandleFunc("/", createSave("/", RequestRoute{
		Get: public(routes.GetEntries),
	}))
	http.HandleFunc("/players/{uuid}", create(RequestRoute{
		Get:    public(routes.GetPlayerData),
		Delete: authenticated(routes.DeletePlayer),
	}))
	http.HandleFunc("/players/{uuid}/data", create(RequestRoute{
		Post: public(routes.UpdatePlayerCustomData),
	}))
	http.HandleFunc("/players/{uuid}/cosmetics/{cosmetic_id}", create(RequestRoute{
		Post:   public(routes.AddPlayerCosmetic),
		Delete: public(routes.RemovePlayerCosmetic),
	}))

	fmt.Printf("Listening on 0.0.0.0:%s\n", routeContext.Config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", routeContext.Config.Port), nil)

	if err != nil {
		panic(err)
	}
}
