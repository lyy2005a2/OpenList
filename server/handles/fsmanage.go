package handles

import (
	"errors"
	"fmt"
	"strings"
	stdpath "path"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/fs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/task"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type MoveCopyReq struct {
	SrcDir       string   `json:"src_dir" form:"src_dir"`
	DstDir       string   `json:"dst_dir" form:"dst_dir"`
	Names        []string `json:"names" form:"names"`
	Overwrite    bool     `json:"overwrite" form:"overwrite"`
	SkipExisting bool     `json:"skip_existing" form:"skip_existing"`
}

func FsMove(c *gin.Context) {
	var req MoveCopyReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Names) == 0 {
		common.ErrorStrResp(c, "文件名不能为空", 400)
		return
	}
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	if !user.CanMove() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	srcDir, err := user.JoinPath(req.SrcDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	srcMeta, err := op.GetNearestMeta(srcDir)
	if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
		common.ErrorResp(c, err, 500, true)
		return
	}
	if !common.CanRead(user, srcMeta, srcDir) {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	dstDir, err := user.JoinPath(req.DstDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	dstMeta, err := op.GetNearestMeta(dstDir)
	if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
		common.ErrorResp(c, err, 500, true)
		return
	}
	if !common.CanWrite(user, dstMeta, dstDir) {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	if !strings.HasSuffix(srcDir, "/") {
		srcDir += "/"
	}
	for i, name := range req.Names {
		srcPath := stdpath.Join(srcDir, name)
		if !strings.HasPrefix(srcPath+"/", srcDir) {
			req.Names[i] = ""
			continue
		}
		if !req.Overwrite {
			base := stdpath.Base(srcPath)
			if base == "." || base == "/" {
				common.ErrorStrResp(c, fmt.Sprintf("无效的文件名 [%s]", name), 400)
				return
			}
			dstPath := stdpath.Join(dstDir, base)
			if res, _ := fs.Get(c.Request.Context(), dstPath, &fs.GetArgs{NoLog: true}); res != nil {
				common.ErrorStrResp(c, fmt.Sprintf("文件 [%s] 已存在", name), 403)
				return
			}
		}
		req.Names[i] = srcPath
	}

	var addedTasks []task.TaskExtensionInfo
	for i, p := range req.Names {
		if p == "" {
			continue
		}
		t, err := fs.Move(c.Request.Context(), p, dstDir, len(req.Names) > i+1)
		if t != nil {
			addedTasks = append(addedTasks, t)
		}
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}

	if len(addedTasks) > 0 {
		common.SuccessResp(c, gin.H{
			"message": fmt.Sprintf("成功创建了 %d 个移动任务", len(addedTasks)),
			"tasks":   getTaskInfos(addedTasks),
		})
	} else {
		common.SuccessResp(c, gin.H{
			"message": "移动操作已立即完成",
		})
	}
}

func FsCopy(c *gin.Context) {
	var req MoveCopyReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Names) == 0 {
		common.ErrorStrResp(c, "文件名不能为空", 400)
		return
	}
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	if !user.CanCopy() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	srcDir, err := user.JoinPath(req.SrcDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	srcMeta, err := op.GetNearestMeta(srcDir)
	if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
		common.ErrorResp(c, err, 500, true)
		return
	}
	if !common.CanRead(user, srcMeta, srcDir) {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	dstDir, err := user.JoinPath(req.DstDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	dstMeta, err := op.GetNearestMeta(dstDir)
	if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
		common.ErrorResp(c, err, 500, true)
		return
	}
	if !common.CanWrite(user, dstMeta, dstDir) {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}

	if !strings.HasSuffix(srcDir, "/") {
		srcDir += "/"
	}
	for i, name := range req.Names {
		srcPath := stdpath.Join(srcDir, name)
		if !strings.HasPrefix(srcPath+"/", srcDir) {
			req.Names[i] = ""
			continue
		}
		req.Names[i] = srcPath
		if !req.Overwrite {
			base := stdpath.Base(srcPath)
			if base == "." || base == "/" {
				common.ErrorStrResp(c, fmt.Sprintf("无效的文件名 [%s]", name), 400)
				return
			}
			dstPath := stdpath.Join(dstDir, base)
			if res, _ := fs.Get(c.Request.Context(), dstPath, &fs.GetArgs{NoLog: true}); res != nil {
				common.ErrorStrResp(c, fmt.Sprintf("文件 [%s] 已存在", name), 403)
				return
			}
		}
	}

	var addedTasks []task.TaskExtensionInfo
	for i, p := range req.Names {
		if p == "" {
			continue
		}
		t, err := fs.Copy(c.Request.Context(), p, dstDir, len(req.Names) > i+1)
		if t != nil {
			addedTasks = append(addedTasks, t)
		}
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}

	if len(addedTasks) > 0 {
		common.SuccessResp(c, gin.H{
			"message": fmt.Sprintf("成功创建了 %d 个复制任务", len(addedTasks)),
			"tasks":   getTaskInfos(addedTasks),
		})
	} else {
		common.SuccessResp(c, gin.H{
			"message": "复制操作已立即完成",
		})
	}
}
