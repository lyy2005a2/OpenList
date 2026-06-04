package handles

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type UpdateSharingReq struct {
	Files       []string `json:"files"`
	Expires     *int64   `json:"expires"`
	Pwd         string   `json:"pwd"`
	Accessed    int      `json:"accessed"`
	MaxAccessed int      `json:"max_accessed"`
	Disabled    bool     `json:"disabled"`
	Sort        model.Sort
	Remark      string `json:"remark"`
	Readme      string `json:"readme"`
	Header      string `json:"header"`
	CreatorName string `json:"creator"`
	Accessed    int    `json:"accessed"`
	ID          string `json:"id"`
	NewID       string `json:"new_id"`
}

var validSharingID = regexp.MustCompile(`^[\w\p{Han}\-]+$`)

func validateSharingID(id string) error {
	if len([]rune(id)) > 64 {
		return errors.New("share id must be at most 64 characters")
	}
	if !validSharingID.MatchString(id) {
		return errors.New("share id can only contain letters, numbers, underscores, hyphens, and CJK characters")
	}
	return nil
}

func UpdateSharing(c *gin.Context) {
	var req UpdateSharingReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Files) == 0 || (len(req.Files) == 1 && req.Files[0] == "") {
		common.ErrorStrResp(c, "must add at least 1 object", 400)
		return
	}
	var user *model.User
	var err error
	reqUser := c.Request.Context().Value(conf.UserKey).(*model.User)
	if reqUser.IsAdmin() && req.CreatorName != "" {
		user, err = op.GetUserByName(req.CreatorName)
		if err != nil {
			common.ErrorStrResp(c, "no such a user", 400)
			return
		}
	} else {
		user = reqUser
		if !user.CanShare() {
			common.ErrorStrResp(c, "权限不足", 403)
			return
		}
	}
	for i, s := range req.Files {
		s = utils.FixAndCleanPath(s)
		req.Files[i] = s
		if !reqUser.IsAdmin() && !strings.HasPrefix(s, user.BasePath) {
			common.ErrorStrResp(c, fmt.Sprintf("permission denied to share path [%s]", s), 500)
			return
		}
	}
	s, err := op.GetSharingById(req.ID)
	if err != nil || (!reqUser.IsAdmin() && s.CreatorId != user.ID) {
		common.ErrorStrResp(c, "sharing not found", 404)
		return
	}
	if reqUser.IsAdmin() && req.CreatorName == "" {
		user = s.Creator
	}
	s.Files = req.Files
	s.Expires = req.Expires
	s.Pwd = req.Pwd
	s.Accessed = req.Accessed
	s.MaxAccessed = req.MaxAccessed
	s.Disabled = req.Disabled
	s.Sort = req.Sort
	s.Header = req.Header
	s.Readme = req.Readme
	s.Remark = req.Remark
	s.Creator = user
	if req.NewID != "" && req.NewID != req.ID {
		if !reqUser.CanCustomizeShareID() {
			common.ErrorStrResp(c, "权限不足", 403)
			return
		}
		if err = validateSharingID(req.NewID); err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		if err = op.UpdateSharingId(s, req.NewID); err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	if err = op.UpdateSharing(s); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c, SharingResp{
			Sharing:     s,
			CreatorName: s.Creator.Username,
			CreatorRole: s.Creator.Role,
		})
	}
}

func CreateSharing(c *gin.Context) {
	var req UpdateSharingReq
	var err error
	if err = c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Files) == 0 || (len(req.Files) == 1 && req.Files[0] == "") {
		common.ErrorStrResp(c, "must add at least 1 object", 400)
		return
	}
	if req.ID != "" {
		if err = validateSharingID(req.ID); err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
	}
	var user *model.User
	reqUser := c.Request.Context().Value(conf.UserKey).(*model.User)
	if reqUser.IsAdmin() && req.CreatorName != "" {
		user, err = op.GetUserByName(req.CreatorName)
		if err != nil {
			common.ErrorStrResp(c, "no such a user", 400)
			return
		}
	} else {
		user = reqUser
		if !user.CanShare() || (!user.CanCustomizeShareID() && req.ID != "") {
			common.ErrorStrResp(c, "权限不足", 403)
			return
		}
	}
	for i, s := range req.Files {
		s = utils.FixAndCleanPath(s)
		req.Files[i] = s
		if !reqUser.IsAdmin() && !strings.HasPrefix(s, user.BasePath) {
			common.ErrorStrResp(c, fmt.Sprintf("permission denied to share path [%s]", s), 500)
			return
		}
	}
	s := &model.Sharing{
		SharingDB: &model.SharingDB{
			ID:          req.ID,
			Expires:     req.Expires,
			Pwd:         req.Pwd,
			Accessed:    req.Accessed,
			MaxAccessed: req.MaxAccessed,
			Disabled:    req.Disabled,
			Sort:        req.Sort,
			Remark:      req.Remark,
			Readme:      req.Readme,
			Header:      req.Header,
		},
		Files:   req.Files,
		Creator: user,
	}
	var id string
	if id, err = op.CreateSharing(s); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		s.ID = id
		common.SuccessResp(c, SharingResp{
			Sharing:     s,
			CreatorName: s.Creator.Username,
			CreatorRole: s.Creator.Role,
		})
	}
}
