package handles

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

func ssoRedirectUri(c *gin.Context, useCompatibility bool, method string) string {
	if useCompatibility {
		return common.GetApiUrl(c) + "/api/auth/" + method
	} else {
		return common.GetApiUrl(c) + "/api/auth/sso_callback" + "?method=" + method
	}
}

func SSOLoginRedirect(c *gin.Context) {
	method := c.Query("method")
	useCompatibility := setting.GetBool(conf.SSOCompatibilityMode)
	enabled := setting.GetBool(conf.SSOLoginEnabled)
	clientId := setting.GetStr(conf.SSOClientId)
	platform := setting.GetStr(conf.SSOLoginPlatform)
	var rUrl string
	if !enabled {
		common.ErrorStrResp(c, "单点登录未启用", 403)
		return
	}
	urlValues := url.Values{}
	if method == "" {
		common.ErrorStrResp(c, "未提供方法", 400)
		return
	}
	redirectUri := ssoRedirectUri(c, useCompatibility, method)
	urlValues.Add("response_type", "code")
	urlValues.Add("redirect_uri", redirectUri)
	urlValues.Add("client_id", clientId)
	switch platform {
	case "Github":
		rUrl = "https://github.com/login/oauth/authorize?"
		urlValues.Add("scope", "read:user")
	case "Microsoft":
		rUrl = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?"
		urlValues.Add("scope", "user.read")
		urlValues.Add("response_mode", "query")
	case "Google":
		rUrl = "https://accounts.google.com/o/oauth2/v2/auth?"
		urlValues.Add("scope", "https://www.googleapis.com/auth/userinfo.profile")
	case "Dingtalk":
		rUrl = "https://login.dingtalk.com/oauth2/auth?"
		urlValues.Add("scope", "openid")
		urlValues.Add("prompt", "consent")
		urlValues.Add("response_type", "code")
	case "Casdoor":
		endpoint := strings.TrimSuffix(setting.GetStr(conf.SSOEndpointName), "/")
		rUrl = endpoint + "/login/oauth/authorize?"
		urlValues.Add("scope", "profile")
		urlValues.Add("state", endpoint)
	case "OIDC":
		oauth2Config, err := GetOIDCClient(c, useCompatibility, redirectUri, method)
		if err != nil {
			common.ErrorStrResp(c, err.Error(), 400)
			return
		}
		state := generateState(clientId, c.ClientIP())
		c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
		return
	default:
		common.ErrorStrResp(c, "无效的平台", 400)
		return
	}
	c.Redirect(302, rUrl+urlValues.Encode())
}
