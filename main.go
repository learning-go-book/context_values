package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/learning-go-book/context_values/identity"
	"github.com/learning-go-book/context_values/tracker"
)

func main() {
	bl := BusinessLogic{
		RequestDecorator: tracker.Request,
		Logger:           tracker.Logger{},
		Remote:           "http://www.example.com/query",
	}

	c := Controller{
		Logic: bl,
	}

	r := chi.NewRouter()
	r.Use(tracker.Middleware, identity.Middleware)
	r.Get("/", c.handleRequest)
	http.ListenAndServe(":3000", r)
}

type RequestDecorator func(*http.Request) *http.Request

type Logger interface {
	Log(context.Context, string)
}
type BusinessLogic struct {
	RequestDecorator RequestDecorator
	Logger           Logger
	Remote           string
}

func (bl BusinessLogic) businessLogic(ctx context.Context, user string, data string) (string, error) {
	bl.Logger.Log(ctx, fmt.Sprintf("starting businessLogic for %s with %s", user, data))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bl.Remote+"?query="+data, nil)
	if err != nil {
		bl.Logger.Log(ctx, fmt.Sprintf("error building remote request: %v", err))
		return "", err
	}
	req = bl.RequestDecorator(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		bl.Logger.Log(ctx, fmt.Sprintf("error making remote request: %v", err))
		return "", err
	}
	body := resp.Body
	defer body.Close()
	out, err := ioutil.ReadAll(body)
	if err != nil {
		bl.Logger.Log(ctx, fmt.Sprintf("error reading response: %v", err))
		return "", err
	}
	return string(out), nil
}

type Logic interface {
	businessLogic(ctx context.Context, user string, data string) (string, error)
}

type Controller struct {
	Logic Logic
}

func (c Controller) handleRequest(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	user, ok := identity.UserFromContext(ctx)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := req.URL.Query().Get("data")
	result, err := c.Logic.businessLogic(ctx, user, data)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}
	rw.Write([]byte(result))
}
