package presetHandlers

import (
	"regexp"
	"fmt"
	"github.com/gin-gonic/gin"
)

func WidthLimitHandler() func(ctx *gin.Context, res *string) *string {
	return func(ctx *gin.Context, res *string) *string {
		r := ""
		//判断是否为max-width-limit-[0-9]+.css, 是的话返回一个css文件
		reg, _ := regexp.Compile(`^/max-width-limit-(\d+).css$`)
		if limit := reg.FindStringSubmatch(ctx.Param("uri")); limit != nil {
			ctx.Header("Content-Type", "text/css")
			ctx.Header("Cache-Control", "max-age=3600000000000")
			ctx.Status(200)
			r = fmt.Sprintf(`p{max-width:%spx!important;margin-left: auto!important;margin-right: auto!important;}`, limit[1])
			return &r
		}
		//给<head></head>里面加一个<link/>样式表
		r = *res
		if limit := ctx.Query("limit"); limit != "" {
			r = addWidthLimit(r, ctx.Query("limit"))
		}
		return &r
	}
}

/*添加p标签宽度限制*/
func addWidthLimit(s, limit string) string {
	r := regexp.MustCompile(`(?i)<head.*>`)
	l := r.FindStringIndex(s)
	LINK := fmt.Sprintf(`<link rel="stylesheet" type="text/css" href="/max-width-limit-%s.css">`, limit)
	//将其加到<head>之后
	return s[:l[1]] + LINK + s[l[1]:]
}
