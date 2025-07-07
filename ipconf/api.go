package ipconf

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/lewyhua/plato/ipconf/domain"
)

type Response struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

func GetIPInfoList(c context.Context, ctx *app.RequestContext) {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		// 打印错误日志
	// 		logger.CtxErrorf(c, "GetIPInfoList.err :%s", err)
	// 		ctx.JSON(consts.StatusInternalServerError, utils.H{"message": err.(error).Error(), "code": 500, "data": nil})
	// 	}
	// }()

	// 1. 构建客户请求信息
	ipConfCtx := domain.BuildIPConfContext(&c, ctx)
	// 2. 进行ip调度
	eds := domain.Dispatch(ipConfCtx)
	// 3. 根据得分取top5返回
	ipConfCtx.AppCtx.JSON(consts.StatusOK, packRes(top5Endpoints(eds)))
}
