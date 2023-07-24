package handler

import (
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"github.com/klintcheng/kim/logger"
)

type LoginHandler struct{}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h LoginHandler) DoSysLogin(ctx him.Context) {
	//log := logger.WithField("func", "DoSysLogin")
	// 序列化
	var session pkt.Session
	if err := ctx.ReadBody(&session); err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	// 检查当前账号是否已经登录在其它地方
	old, err := ctx.GetLocation(session.Account, "")
	if err != nil && err != him.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	if old != nil {
		// 通知用户下线
		_ = ctx.Dispatch(&pkt.KickoutNotify{ChannelId: old.ChannelId})
		return
	}
	// 添加到会话管理器
	err = ctx.Add(&session)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	// 返回一个登录成功的消息
	resp := &pkt.LoginResp{
		ChannelId: session.ChannelId,
	}
	_ = ctx.Resp(pkt.Status_Success, resp)
}

func (h LoginHandler) DoSysLogout(ctx him.Context) {
	logger.WithField("func", "DoSysLogout").Info("do Logout of %s %S ", ctx.Session().GetChannelId(), ctx.Session().GetAccount())

	err := ctx.Delete(ctx.Session().GetAccount(), ctx.Session().GetChannelId())
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}
