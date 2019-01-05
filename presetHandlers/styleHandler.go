package presetHandlers

import (
	"regexp"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func StyleHandler(src string) func(*gin.Context, *http.Response, *string) *string {
	return func(ctx *gin.Context, res *http.Response, body *string) *string {
		r := *body
		if strings.HasSuffix(res.Request.URL.Host, "shu.edu.cn") {
			r = addCSSLink(r, src)
		}
		return &r
	}
}

/*添加<link>*/
func addStyle(s string, src string) string {
	r := regexp.MustCompile(`(?i)<head.*?>`)
	if l := r.FindStringIndex(s); len(l) > 0 {
		LINK := `<link rel="stylesheet" type="text/css" href="` + src + `">`
		//将其加到<head>之后
		return s[:l[1]] + LINK + s[l[1]:]
	}
	return s
}
