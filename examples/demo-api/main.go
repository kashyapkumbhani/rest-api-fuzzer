package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

type productRequest struct {
	Name  string   `json:"name"`
	Price float64  `json:"price"`
	Tags  []string `json:"tags"`
}

func main() {
	handler := func(ctx *fasthttp.RequestCtx) {
		switch {
		case string(ctx.Path()) == "/health" && ctx.IsGet():
			writeJSON(ctx, fasthttp.StatusOK, map[string]string{"status": "ok"})
		case string(ctx.Path()) == "/products" && ctx.IsPost():
			createProduct(ctx)
		case strings.HasPrefix(string(ctx.Path()), "/products/") && ctx.IsGet():
			getProduct(ctx)
		default:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
	}

	if err := fasthttp.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}

func createProduct(ctx *fasthttp.RequestCtx) {
	var req productRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if req.Name == "" || req.Price < 0 {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid product"})
		return
	}
	if strings.Contains(req.Name, "' OR '1'='1") {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "query builder exploded"})
		return
	}
	if len(req.Tags) > 2 {
		time.Sleep(900 * time.Millisecond)
	}
	writeJSON(ctx, fasthttp.StatusCreated, map[string]any{
		"id":    "00000000-0000-4000-8000-000000000001",
		"name":  req.Name,
		"price": req.Price,
	})
}

func getProduct(ctx *fasthttp.RequestCtx) {
	writeJSON(ctx, fasthttp.StatusOK, map[string]any{
		"id":    strings.TrimPrefix(string(ctx.Path()), "/products/"),
		"name":  "demo",
		"price": "19.99",
	})
}

func writeJSON(ctx *fasthttp.RequestCtx, status int, body any) {
	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType("application/json")
	encoded, err := json.Marshal(body)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetBody(encoded)
}
