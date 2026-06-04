package handles

import (
	"github.com/OpenListTeam/OpenList/v4/drivers/_115_open"
	_123 "github.com/OpenListTeam/OpenList/v4/drivers/_123"
	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/internal/tool"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type Set115OpenReq struct {
	TempDir string `json:"temp_dir" form:"temp_dir"`
}

func Set115Open(c *gin.Context) {
	var req Set115OpenReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if req.TempDir != "" {
		storage, _, err := op.GetStorageAndActualPath(req.TempDir)
		if err != nil {
			common.ErrorStrResp(c, "存储不存在", 400)
			return
		}
		if storage.Config().CheckStatus && storage.GetStorage().Status != op.WORK {
			common.ErrorStrResp(c, "storage not init: "+storage.GetStorage().Status, 400)
			return
		}
		if _, ok := storage.(*_115_open.Open115); !ok {
			common.ErrorStrResp(c, "不支持的离线下载存储驱动，仅支持 115 Open", 400)
			return
		}
	}
	items := []model.SettingItem{
		{Key: conf.Pan115OpenTempDir, Value: req.TempDir, Type: conf.TypeString, Group: model.OFFLINE_DOWNLOAD, Flag: model.PRIVATE},
	}
	if err := op.SaveSettingItems(items); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	_tool, err := tool.Tools.Get("115 Open")
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if _, err := _tool.Init(); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, "ok")
}

type Set123PanReq struct {
	TempDir string `json:"temp_dir" form:"temp_dir"`
}

func Set123Pan(c *gin.Context) {
	var req Set123PanReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if req.TempDir != "" {
		storage, _, err := op.GetStorageAndActualPath(req.TempDir)
		if err != nil {
			common.ErrorStrResp(c, "存储不存在", 400)
			return
		}
		if storage.Config().CheckStatus && storage.GetStorage().Status != op.WORK {
			common.ErrorStrResp(c, "storage not init: "+storage.GetStorage().Status, 400)
			return
		}
		if _, ok := storage.(*_123.Pan123); !ok {
			common.ErrorStrResp(c, "不支持的离线下载存储驱动，仅支持 123Pan", 400)
			return
		}
	}
	items := []model.SettingItem{
		{Key: conf.Pan123TempDir, Value: req.TempDir, Type: conf.TypeString, Group: model.OFFLINE_DOWNLOAD, Flag: model.PRIVATE},
	}
	if err := op.SaveSettingItems(items); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	_tool, err := tool.Tools.Get("123Pan")
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if _, err := _tool.Init(); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, "ok")
}

type Set123OpenReq struct {
	TempDir     string `json:"temp_dir" form:"temp_dir"`
	CallbackUrl string `json:"callback_url" form:"callback_url"`
}

func Set123Open(c *gin.Context) {
	var req Set123OpenReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if req.TempDir != "" {
		storage, _, err := op.GetStorageAndActualPath(req.TempDir)
		if err != nil {
			common.ErrorStrResp(c, "存储不存在", 400)
			return
		}
		if storage.Config().CheckStatus && storage.GetStorage().Status != op.WORK {
			common.ErrorStrResp(c, "storage not init: "+storage.GetStorage().Status, 400)
			return
		}
		if _, ok := storage.(*_123.Pan123Open); !ok {
			common.ErrorStrResp(c, "不支持的离线下载存储驱动，仅支持 123 Open", 400)
			return
		}
	}
	items := []model.SettingItem{
		{Key: conf.Pan123OpenTempDir, Value: req.TempDir, Type: conf.TypeString, Group: model.OFFLINE_DOWNLOAD, Flag: model.PRIVATE},
		{Key: conf.Pan123OpenCallbackUrl, Value: req.CallbackUrl, Type: conf.TypeString, Group: model.OFFLINE_DOWNLOAD, Flag: model.PRIVATE},
	}
	if err := op.SaveSettingItems(items); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	_tool, err := tool.Tools.Get("123 Open")
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if _, err := _tool.Init(); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, "ok")
}
