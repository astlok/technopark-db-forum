package httputils

import (
	"encoding/json"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"log"
)

func Respond(ctx *fasthttp.RequestCtx, code int, data easyjson.Marshaler) {
	ctx.SetStatusCode(code)
	ctx.Response.Header.Set("Content-Type", "application/json")
	if data != nil {
		_, err := easyjson.MarshalToWriter(data, ctx)
		//err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Print(err, data)
			return
		}
	}
}

func RespondErr(ctx *fasthttp.RequestCtx, code int, data interface{}) {
	ctx.SetStatusCode(code)
	ctx.Response.Header.Set("Content-Type", "application/json")
	if data != nil {
		err := json.NewEncoder(ctx).Encode(data)
		if err != nil {
			log.Print(err, data)
			return
		}
	}
}
