package shumsgHandlers

import (
	"regexp"
	"github.com/gin-gonic/gin"
)

func SeoNormalizeHandler() func(ctx *gin.Context, res *string) *string {
	return func(ctx *gin.Context, res *string) *string {
		r := addSeoNormalize(*res)
		return &r
	}
}

/*添加seoNormalize.css*/
func addSeoNormalize(s string) string {
	r := regexp.MustCompile(`(?i)<head.*>`)
	l := r.FindStringIndex(s)
	LINK := `<link rel="stylesheet" type="text/css" href="https://shumsg.cn/static/seoNormalize.css">`
	//将其加到<head>之后
	return s[:l[1]] + LINK + s[l[1]:]
}
