package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"strings"
	"compress/gzip"
	"bytes"
)

const InitialStatus = 1

/*
当需要终止的时候请调用ctx.Abort()、ctx.AbortWithStatus()等函数，但请不要在任何情况下调用ctx.Next()
*/
type ResponseHandler struct {
	handler           func(*gin.Context, *http.Response)
	deferCallbackFunc *func()
}
type ResponseHandlersChain []ResponseHandler

func (chain *ResponseHandlersChain) responseUse(middleware ...ResponseHandler) {
	*chain = append(*chain, middleware...)
}

func (chain *ResponseHandler) deferCallback(callback func()) {
	chain.deferCallbackFunc = &callback
}

/*
目前版本，如果response经过压缩，只支持gzip方式解压，并且最终会对内容进行gzip压缩
*/
type responseBodyHandlerFunc func(*gin.Context, *string) *string

func (chain *ResponseHandlersChain) responseBodyUse(handlers ...responseBodyHandlerFunc) {
	var callback func() = nil
	hdl := ResponseHandler{handler: func(ctx *gin.Context, res *http.Response) {
		//siteUrl := siteUrl(getHostFromUrl(res.Request.URL.String(), true))
		buf := new(bytes.Buffer)
		var s string
		if ct := res.Header.Get("Content-Type"); strings.Index(strings.ToLower(ct), "text/html") > -1 {
			if ce := res.Header.Get("Content-Encoding"); strings.ToLower(ce) == "gzip" {
				//gzip解压
				r, _ := gzip.NewReader(res.Body)
				defer r.Close()
				buf.ReadFrom(r)
			} else {
				buf.ReadFrom(res.Body)
			}
			s = buf.String()
			//调用handlers之前先将status置特殊值InitialStatus=1，以此监测status是否经handlers发生了变化
			ctx.Status(InitialStatus)
			//调用handlers
			for _, handler := range handlers {
				if r := handler(ctx, &s); r != nil {
					s = *r
				}
			}
			//如果是超过1kb的html，则进行gzip压缩
			var length int
			if ct := res.Header.Get("Content-Type"); len(s) >= 1024 && strings.Index(strings.ToLower(ct), "text/html") > -1 {
				buf.Reset()
				gw := gzip.NewWriter(buf)
				defer gw.Close()
				gw.Write([]byte(s))
				//写回buf
				gw.Flush()
				s = buf.String()
				//补足Content-Encoding为gzip
				if ce := res.Header.Get("Content-Encoding"); ce != "gzip" {
					ctx.Header("Content-Encoding", "gzip")
				}
				length = buf.Len()
			} else {
				length = len(s)
			}
			ctx.Header("Content-Length", fmt.Sprint(length))
		} else {
			buf.ReadFrom(res.Body)
			s = buf.String()
		}
		//如果有handler对status进行了改变，则使用改变后的status，否则沿用response的status
		if status := ctx.Writer.Status(); status == InitialStatus {
			callback = func() { ctx.String(res.StatusCode, s) }
		} else {
			callback = func() { ctx.String(status, s) }
		}
	}, deferCallbackFunc: &callback}
	chain.responseUse(hdl)
}
