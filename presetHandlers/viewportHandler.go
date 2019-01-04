package presetHandlers

import (
	"regexp"
	"github.com/gin-gonic/gin"
)

func ViewportHandler() func(ctx *gin.Context, res *string) *string {
	return func(ctx *gin.Context, res *string) *string {
		r := addViewportSupport(*res)
		return &r
	}
}

func addViewportSupport(HTML string) string {
	const META = `<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0">`
	reg, _ := regexp.Compile(`<head.*?>[\S\s]*?(<meta.*?name="viewport".*?>)`)
	loc := reg.FindStringSubmatchIndex(HTML)
	var ns string
	if len(loc) > 0 {
		//存在viewport标签，将其替换
		ns = HTML[:loc[2]] + META + HTML[loc[3]:]
	} else {
		reg, _ := regexp.Compile(`<head.*?>`)
		if loc := reg.FindStringIndex(HTML); len(loc) > 0 {
			//不存在viewport标签，将其加入head
			ns = HTML[:loc[1]] + "\n  " + META + HTML[loc[1]:]
		}
	}
	return ns
}
