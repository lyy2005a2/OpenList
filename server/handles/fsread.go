package handles

import (
	"errors"
	"url"
	"strconv"
	"strings"
	stdpath "path"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/fs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type ListReq struct {
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password" default:"/"`
	common.PageReq
	Refresh bool `json:"refresh" form:"refresh"`
}

type DirReq struct {
	Path      string `json:"path" form:"path"`
	Password  string `json:"password" form:"password" default:"/"`
	ForceRoot bool   `json:"force_root" form:"force_root"`
}

type FsListResp struct {
	Content            []interface{} `json:"content"`
	Total              int64         `json:"total"`
	Readme             string        `json:"readme"`
	Header             string        `json:"header"`
	Write              bool          `json:"write"`
	WriteContentBypass bool          `json:"write_content_bypass"`
	Provider           string        `json:"provider"`
	DirectUploadTools  []string      `json:"direct_upload_tools"`
}

func FsList(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	if user.IsGuest() && !op.CanGuestUser() {
		common.ErrorStrResp(c, "游客用户已禁用，请登录", 401)
		return
	}
	var req ListReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	FsList(c, &req, user)
}

func FsList(c *gin.Context, req *ListReq, user *model.User) {
	reqPath, err := user.JoinPath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	meta, err := op.GetNearestMeta(reqPath)
	if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.GinAppendValues(c, conf.MetaKey, meta)
	if !common.CanAccess(user, meta, reqPath, req.Password) {
		common.ErrorStrResp(c, "密码错误或无权限访问", 403)
		return
	}
	canWriteContentAtPath := common.CanWrite(user, meta, reqPath) && (user.CanWriteContent() || common.CanWriteContentBypassUserPerms(meta, reqPath))
	if req.Refresh && !canWriteContentAtPath {
		common.ErrorStrResp(c, "刷新无权限", 403)
		return
	}
	objs, err := fs.List(c.Request.Context(), reqPath, &fs.ListArgs{
		Refresh:            req.Refresh,
		WithStorageDetails: !user.IsGuest() && !setting.GetBool(conf.HideStorageDetails),
	})
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	total, objs := pagination(objs, &req.PageReq)
	provider := "unknown"
	var directUploadTools []string
	if canWriteContentAtPath {
		if storage, err := fs.GetStorage(reqPath, &fs.GetStoragesArgs{}); err == nil {
			directUploadTools = op.GetDirectUploadTools(storage)
		}
	}
	common.SuccessResp(c, FsListResp{
		Content:            toObjsResp(objs, reqPath, isEncrypt(meta, reqPath)),
		Total:              int64(total),
		Readme:             getReadme(meta, reqPath),
		Header:             getHeader(meta, reqPath),
		Write:              common.CanWrite(user, meta, reqPath),
		WriteContentBypass: common.CanWriteContentBypassUserPerms(meta, reqPath),
		Provider:           provider,
		DirectUploadTools:  directUploadTools,
	})
}

func FsDirs(c *gin.Context) {
	var req DirReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	reqPath := req.Path
	if req.ForceRoot {
		if !user.IsAdmin() {
			common.ErrorStrResp(c, "权限不足", 403)
			return
		}
	} else {
		tmp, err := user.JoinPath(req.Path)
		if err != nil {
			common.ErrorResp(c, err, 403)
			return
		}
		reqPath = tmp
	}
	meta, err := op.GetNearestMeta(reqPath)
	if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.GinAppendValues(c, conf.MetaKey, meta)
	if !common.CanAccess(user, meta, reqPath, req.Password) {
		common.ErrorStrResp(c, "密码错误或无权限访问", 403)
		return
	}
	var dirs []interface{}
	objs, err := fs.List(c.Request.Context(), reqPath, &fs.ListArgs{})
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	for _, obj := range objs {
		if obj.IsDir() {
			dirs = append(dirs, gin.H{
				"name": obj.GetName(),
				"path": reqPath + "/" + obj.GetName(),
			})
		}
	}
	common.SuccessResp(c, dirs)
}
