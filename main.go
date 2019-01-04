package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"bytes"
	"time"
	"strings"
	"regexp"
	"fmt"
	"encoding/json"
	"errors"
	"log"
	"github.com/gin-contrib/cors"
	"github.com/Actooors/iframeForward/presetHandlers"
	"github.com/Actooors/iframeForward/shumsgHandlers"
)

var selfHost = []string{"0.0.0.0:8090", "api.mzz.pub:8090", "192.168.50.111:8090", "proxy.shumsg.cn"}
var frontHost = []string{"api.mzz.pub:8000", "shumsg.cn"}

const FirstRequestPath = "/getforward/get"
const ApiRoot = "http://api.mzz.pub:8188/api"

type siteUrl string

var responseHandlersChain ResponseHandlersChain

func main() {
	responseHandlersChain.responseBodyUse(
		presetHandlers.ViewportHandler(),
		presetHandlers.WidthLimitHandler(),
		shumsgHandlers.SeoNormalizeHandler(),
	)

	router := gin.Default()
	trustHosts := make([]string, len(frontHost)*2)
	for i, t := range frontHost {
		trustHosts[2*i] = "http://" + t
		trustHosts[2*i+1] = "https://" + t
	}
	fmt.Println(trustHosts)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = trustHosts
	router.Use(cors.New(corsConfig))
	router.Use(gin.Recovery())
	router.Any("*uri", anyForward)
	router.Run(":8090")
}

/*
	转发所有请求到cookie标识的站点
*/
func anyForward(ctx *gin.Context) {
	url2 := ctx.Param("uri")
	type cookieSaver struct {
		value  string
		maxAge int
		domain string
	}
	var cs *cookieSaver = nil
	var firstAcess = false
	/*
		首次访问该站点，留下1个小时的cookie，实现具有一定粘性的反向代理
	*/
	if ctx.Request.URL.Path == FirstRequestPath && strings.HasPrefix(ctx.Request.URL.RawQuery, "__url=") {
		firstAcess = true
		url2 = ctx.Query("__url")
		host := getHostFromUrl(url2, true)
		domain := ctx.Request.Host
		if index := strings.Index(ctx.Request.Host, ":"); index > 0 {
			domain = domain[:index]
		}
		//fmt.Print(ctx.Request.Host)
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
		//reg, _ := regexp.Compile(FirstRequestPath + `\?.*url=__https?`)
		refer := ctx.Request.Referer()
		var site string
		var err error
		//直接从cookie取
		site, err = ctx.Cookie("__forward_site")
		//cookie没有，尝试从refer取

		if err != nil {
			log.Println("cookie中没有site，从refer取得: ", site)
			//先尝试是否refer的是/getForward/get接口
			re, _ := regexp.Compile(`.*` + FirstRequestPath + `\?__url=(.*)`)
			result := re.FindStringSubmatch(refer)
			if len(result) >= 2 {
				site = getHostFromUrl(result[1], true)
			} else {
				//如果不是，尝试直接引用refer的site
				urlFromRefer := getHostFromUrl(refer, true)
				site = getHostFromUrl(urlFromRefer, true)
			}
		}
		url2 = site + url2
		h := getHostFromUrl(site, false)
		for _, self := range selfHost {
			if h == self {
				ctx.Status(503)
				log.Println("error: site与本站相同", ctx.Request.Header)
				return
			}
		}
		//log.Println("合成url: ", url2, getHostFromUrl(site, false), selfHost)
	}
	//开始转发请求
	raw, err := ctx.GetRawData()
	if err != nil {
		handleError(ctx, err)
		return
	}
	request, err := http.NewRequest(ctx.Request.Method, url2, bytes.NewReader(raw))
	if err != nil {
		handleError(ctx, err)
		return
	}
	defer request.Body.Close()
	request.Header = ctx.Request.Header
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		handleError(ctx, err)
		return
	}
	defer res.Body.Close()
	supportIframe := true
	siteUrl := siteUrl(getHostFromUrl(url2, true))
	//将友好的response头原原本本添加回去
	for k, v := range res.Header {
		//log.Println("here: ", k, v)
		switch k {
		case "Access-Control-Allow-Origin",
			"Access-Control-Request-Method",
			"Host":
			continue
			//该站点由于有X-Frame-Options首部，因此不支持iframe，我们在数据库对它进行记录
		case "X-Frame-Options":
			if firstAcess {
				go func() {
					err := siteUrl.changeSupportIframeSite(false)
					if err != nil {
						fmt.Println("* When changeSupportIframeSite, ", err)
					}
				}()
				supportIframe = false
			}
			continue
		}
		for _, val := range v {
			ctx.Header(k, val)
		}
	}
	if firstAcess && supportIframe {
		go func() {
			err := siteUrl.changeSupportIframeSite(true)
			if err != nil {
				fmt.Println("* When changeSupportIframeSite, ", err)
			}
		}()
	}
	//将host改为目标域名，以防403
	ctx.Header("Host", getHostFromUrl(url2, false))
	if cs != nil {
		ctx.SetCookie("__forward_site", cs.value, cs.maxAge, "/", cs.domain, false, true)
	}
	//调用responseHandlersChain上的Handlers
	for _, r := range responseHandlersChain {
		r.handler(ctx, res)
		if ctx.IsAborted() {
			return
		}
	}
	for _, r := range responseHandlersChain {
		if callback := r.deferCallbackFunc; callback != nil {
			(*callback)()
		}
	}
	//没有什么callback对内容进行了写操作，就直接将response的body返回
	if !ctx.Writer.Written() {
		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)
		ctx.Data(res.StatusCode, res.Header.Get("Content-Type"), buf.Bytes())
	}
}

func isCompleteURL(url string) bool {
	ok, err := regexp.MatchString(`(?i)^https?://`, url)
	if err != nil {
		return false
	}
	return ok
}

func getHostFromUrl(url string, includeProtocol bool) (host string) {
	t := strings.Index(url, "//")
	if t == -1 {
		//令t+2=0
		t = -2
	}
	host = url[t+2:]
	e := strings.Index(host, "/")
	if e == -1 {
		e = len(host)
	}
	if includeProtocol {
		host = url[:t+2] + host[:e]
	} else {
		host = host[:e]
	}
	return
}

func (str *siteUrl) changeSupportIframeSite(support bool) (error) {
	params := make(map[string]interface{})
	params["host"] = str
	params["support"] = support
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", ApiRoot+"/common/newIframe", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer request.Body.Close()
	request.Header.Set("Accept", "application/json, text/plain, */*")
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	var res struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return errors.New(buf.String() + " | " + err.Error())
	}
	if res.Code == "FAILED" {
		return errors.New(res.Message)
	}
	return nil
}
func handleError(ctx *gin.Context, err error) {
	ctx.Status(503)
	fmt.Println(err)
}
