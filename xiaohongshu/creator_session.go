package xiaohongshu

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
	"github.com/xpzouying/xiaohongshu-mcp/humanize"
)

const (
	creatorHostXHS     = "creator.xiaohongshu.com"
	creatorHostRednote = "creator.rednote.com"

	creatorSessionCookie  = "galaxy_creator_session_id"
	creatorSessionTimeout = 15 * time.Second
)

// ensureCreatorSession 浏览器里没有创作者中心会话时，先打开创作者中心首页，等会话建立。
// 返回本账号所属的创作者中心域名，发布页要用同一个域名。
func ensureCreatorSession(page *rod.Page) (string, error) {
	cks, _ := page.Browser().GetCookies()
	host := resolveCreatorHost(os.Getenv("XHS_CREATOR_HOST"), cks)

	if hasCreatorSession(cks) {
		return host, nil
	}

	if err := page.Navigate(creatorHomeURL(host)); err != nil {
		return "", errors.Wrap(err, "打开创作者中心失败")
	}

	deadline := time.Now().Add(creatorSessionTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		cks, _ = page.Browser().GetCookies()
		if hasCreatorSession(cks) || onCreatorHome(page, host) {
			humanize.Delay(page.GetContext(), humanize.AfterNavigate)
			return host, nil
		}
	}
	return "", errors.New("创作者中心登录失效，请重新扫码登录")
}

// resolveCreatorHost 选创作者中心域名：环境变量优先；
// 海外手机号注册的 rednote 账号，登录态在 .rednote.com 下，走 creator.rednote.com。
func resolveCreatorHost(envHost string, cks []*proto.NetworkCookie) string {
	if envHost != "" {
		return envHost
	}
	if isRednoteAccount(cks) {
		return creatorHostRednote
	}
	return creatorHostXHS
}

// isRednoteAccount 主站登录态 web_session 只在 rednote.com 下时，判定为 rednote 账号
func isRednoteAccount(cks []*proto.NetworkCookie) bool {
	rednote, xhs := false, false
	for _, c := range cks {
		if c.Name != "web_session" {
			continue
		}
		switch {
		case strings.HasSuffix(c.Domain, "rednote.com"):
			rednote = true
		case strings.HasSuffix(c.Domain, "xiaohongshu.com"):
			xhs = true
		}
	}
	return rednote && !xhs
}

func hasCreatorSession(cks []*proto.NetworkCookie) bool {
	for _, c := range cks {
		if c.Name == creatorSessionCookie || strings.HasPrefix(c.Name, "access-token-creator") {
			return true
		}
	}
	return false
}

// onCreatorHome 已停在创作者中心的登录后页面（/new/...），说明会话已建立
func onCreatorHome(page *rod.Page, host string) bool {
	info, err := page.Info()
	if err != nil {
		return false
	}
	u, err := url.Parse(info.URL)
	if err != nil {
		return false
	}
	return u.Hostname() == host && strings.HasPrefix(u.Path, "/new/")
}

func creatorHomeURL(host string) string {
	return "https://" + host + "/?source=official"
}

func creatorPublishURL(host string) string {
	return "https://" + host + "/publish/publish?source=official"
}
