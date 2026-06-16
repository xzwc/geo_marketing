package publisher

import "testing"

func TestAdaptContent(t *testing.T) {
	cases := []struct {
		name, content, src, kind, wantContains string
	}{
		{"md->html 标题", "# 标题", FormatMarkdown, KindHTML, "<h1"},
		{"md->html 加粗", "**粗**", FormatMarkdown, KindHTML, "<strong>"},
		{"md->html 列表", "- a\n- b", FormatMarkdown, KindHTML, "<li>"},
		{"md->html 表格(GFM)", "| a | b |\n|---|---|\n| 1 | 2 |", FormatMarkdown, KindHTML, "<table>"},
		{"md->markdown 原样", "# 标题", FormatMarkdown, KindMarkdown, "# 标题"},
		{"plain->html 段落", "第一段\n\n第二段", FormatPlain, KindHTML, "<p>第一段</p>"},
		{"plain->html 转义", "a<b>c", FormatPlain, KindHTML, "&lt;b&gt;"},
		{"plain->html 换行", "l1\nl2", FormatPlain, KindHTML, "<br>"},
		{"plain->markdown 转义井号", "# 不是标题", FormatPlain, KindMarkdown, "\\#"},
		{"html遗留->html 透传", "<p>x</p>", FormatHTML, KindHTML, "<p>x</p>"},
		{"空格式->html 透传", "<p>y</p>", "", KindHTML, "<p>y</p>"},
		{"空kind默认html: md转换", "# h", FormatMarkdown, "", "<h1"},
	}
	for _, c := range cases {
		got := AdaptContent(c.content, c.src, c.kind)
		if !contains(got, c.wantContains) {
			t.Errorf("[%s] AdaptContent(%q,%q,%q)=%q, 期望包含 %q", c.name, c.content, c.src, c.kind, got, c.wantContains)
		}
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
