package handler

import (
	"errors"
	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"time"
)

var ErrNoDestination = errors.New("dest is empty")

type ChatHandler struct {
}

// DoUserTalk 单聊逻辑
func (h *ChatHandler) DoUserTalk(ctx him.Context) {
	// validate
	if ctx.Header().Dest == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrNoDestination)
		return
	}
	// 解包
	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, ErrNoDestination)
		return
	}
	// 接受方寻址
	dest := ctx.Header().GetDest()
	loc, err := ctx.GetLocation(dest, "")
	if err != nil && err != him.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, ErrNoDestination)
		return
	}
	// 保存离线信息
	sendTime := time.Now().UnixNano()
	// TODO: mysql

	// 接收方在线，发送消息
	if loc != nil {
		if err = ctx.Dispatch(&pkt.MessagePush{
			// TODO
		}, loc); err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}

	}
	// 5. 返回一条resp消息
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResp{
		//MessageId: msgId,
		SendTime: sendTime,
	})
}

func (h *ChatHandler) DoGroupTalk(ctx him.Context) {
	if ctx.Header().GetDest() == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrNoDestination)
		return
	}
	// 解包
	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}
	// 群聊的dest是群ID
	group := ctx.Header().GetDest()
	sendTime := time.Now().UnixNano()

	// TODO: 保存离线消息
	// TODO: 读取群成员列表
	members := make([]string, 5)

	// 批量寻址
	locs, err := ctx.GetLocations(members...)
	if err != nil {
		ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// 批量推送消息给成员

}
