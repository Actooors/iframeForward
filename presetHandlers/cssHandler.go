package presetHandlers

import (
	"regexp"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CSSLinkHandler(src string) func(*gin.Context, *http.Response, *string) *string {
	return func(ctx *gin.Context, res *http.Response, body *string) *string {
		r := *body
		if strings.HasSuffix(res.Request.URL.Host, "shu.edu.cn") {
			r = addCSSLink(r, src)
		}
		return &r
	}
}

func StyleHandler(styleContent string) func(*gin.Context, *http.Response, *string) *string {
	return func(ctx *gin.Context, res *http.Response, body *string) *string {
		r := *body
		if strings.HasSuffix(res.Request.URL.Host, "shu.edu.cn") {
			r = addStyle(r, styleContent)
		}
		return &r
	}
}

/*添加<link>*/
func addCSSLink(s string, src string) string {
	r := regexp.MustCompile(`(?i)<head.*?>`)
	if l := r.FindStringIndex(s); len(l) > 0 {
		LINK := `<link rel="stylesheet" type="text/css" href="` + src + `">`
		//将其加到<head>之后
		return s[:l[1]] + LINK + s[l[1]:]
	}
	return s
}

/*添加<style></style>*/
func addStyle(s string, styleContent string) string {
	r := regexp.MustCompile(`(?i)<head.*?>`)
	if l := r.FindStringIndex(s); len(l) > 0 {
		LINK := `<style>` + styleContent + `</style>`
		//将其加到<head>之后
		return s[:l[1]] + LINK + s[l[1]:]
	}
	return s
}
