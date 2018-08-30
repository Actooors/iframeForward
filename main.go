package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"net/http"
	"bytes"
	"time"
	"strings"
	"regexp"
	"fmt"
)

const FirstRequestPath = "/getforward/get"

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(gin.Recovery())
	//router.GET("/getforward/get", getForward)
	router.Any("*url", anyForward)
	router.Run(":8090")
}

/*
	转发所有请求到cookie标识的站点
*/
func anyForward(ctx *gin.Context) {
	url2 := ctx.Param("url")
	type cookieSaver struct {
		value  string
		maxAge int
		domain string
	}
	var cs *cookieSaver = nil
	/*
		首次访问该站点，留下1个小时的cookie，实现具有一定粘性的反向代理
	*/
	if strings.ToLower(url2) == FirstRequestPath {
		url2 = ctx.Query("url")
		host := getHostFromUrl(url2)
		domain := ctx.Request.Host
		if index := strings.Index(ctx.Request.Host, ":"); index > 0 {
			domain = domain[:index]
		}
		fmt.Print(ctx.Request.Host)
		cs = &cookieSaver{
			value:  host,
			maxAge: int(time.Hour),
			domain: domain,
		}
	}
	//是否是完整的url
	ok := isCompleteURL(url2)
	//并非完整的url
	if !ok {
		//reg, _ := regexp.Compile(FirstRequestPath + `\?.*url=https?`)
		refer := ctx.Request.Referer()
		var site string
		var err error
		//直接从cookie取
		site, err = ctx.Cookie("__forward_site")
		//cookie没有，尝试从refer取
		if urlFromRefer := getHostFromUrl(refer); err != nil && isCompleteURL(urlFromRefer) {
			site = getHostFromUrl(urlFromRefer)
		}
		url2 = site + url2
	}
	raw, err := ctx.GetRawData()
	if err != nil {
		ctx.Status(500)
		return
	}
	request, err := http.NewRequest(ctx.Request.Method, url2, bytes.NewReader(raw))
	if err != nil {
		ctx.Status(500)
		return
	}
	request.Header = ctx.Request.Header
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		ctx.Status(500)
		return
	}
	//将友好的response头原原本本添加回去
	for k, v := range res.Header {
		switch k {
		case "X-Frame-Options", "Access-Control-Allow-Origin", "Access-Control-Request-Method":
			continue
		}
		for _, val := range v {
			ctx.Header(k, val)
		}
	}
	//加上跨域友好response头
	ctx.Header("Access-Control-Allow-Origin", "*")
	if cs != nil {
		ctx.SetCookie("__forward_site", cs.value, cs.maxAge, "/", cs.domain, false, false)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	s := buf.String()
	ctx.String(200, s)
}

func isCompleteURL(url string) bool {
	ok, err := regexp.MatchString(`^https?://`, url)
	if err != nil {
		return false
	}
	return ok
}

func getHostFromUrl(url string) (host string) {
	t := strings.Index(url, "//")
	if t == -1 {
		t = 0
	}
	host = url[t+2:]
	e := strings.Index(host, "/")
	if e == -1 {
		e = len(host)
	}
	host = url[:t+2] + host[:e]
	return
}
