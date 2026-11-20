package xiaohongshu

import (
	"testing"

	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/assert"
)

func ck(name, domain string) *proto.NetworkCookie {
	return &proto.NetworkCookie{Name: name, Domain: domain}
}

// TestResolveCreatorHost 校验创作者中心域名选择：环境变量优先，rednote 账号走 creator.rednote.com
func TestResolveCreatorHost(t *testing.T) {
	tests := []struct {
		name string
		env  string
		cks  []*proto.NetworkCookie
		want string
	}{
		{name: "无cookie默认小红书", want: creatorHostXHS},
		{name: "小红书账号", cks: []*proto.NetworkCookie{ck("web_session", ".xiaohongshu.com")}, want: creatorHostXHS},
		{name: "rednote账号", cks: []*proto.NetworkCookie{ck("acw_tc", "www.xiaohongshu.com"), ck("web_session", ".rednote.com")}, want: creatorHostRednote},
		{name: "两边都有登录态按小红书", cks: []*proto.NetworkCookie{ck("web_session", ".rednote.com"), ck("web_session", ".xiaohongshu.com")}, want: creatorHostXHS},
		{name: "环境变量优先", env: "creator.example.com", cks: []*proto.NetworkCookie{ck("web_session", ".rednote.com")}, want: "creator.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveCreatorHost(tt.env, tt.cks))
		})
	}
}

// TestHasCreatorSession 校验创作者中心会话 cookie 识别
func TestHasCreatorSession(t *testing.T) {
	assert.False(t, hasCreatorSession([]*proto.NetworkCookie{ck("web_session", ".rednote.com")}))
	assert.True(t, hasCreatorSession([]*proto.NetworkCookie{ck("galaxy_creator_session_id", ".rednote.com")}))
	assert.True(t, hasCreatorSession([]*proto.NetworkCookie{ck("access-token-creator.rednote.com", ".rednote.com")}))
}

func TestCreatorURLs(t *testing.T) {
	assert.Equal(t, "https://creator.rednote.com/publish/publish?source=official", creatorPublishURL(creatorHostRednote))
	assert.Equal(t, "https://creator.xiaohongshu.com/?source=official", creatorHomeURL(creatorHostXHS))
}
