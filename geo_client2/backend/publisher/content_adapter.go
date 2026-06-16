package publisher

import (
	"bytes"
	"html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// 用户侧内容源格式
const (
	FormatHTML     = "html"
	FormatMarkdown = "markdown"
	FormatPlain    = "plain"
)

// 平台编辑器期望的目标类型(声明在各 flow 的 contentKind 字段)
const (
	KindHTML     = "html"
	KindMarkdown = "markdown"
)

var mdConverter = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	// 允许用户在 Markdown 中内联的原始 HTML 透传(发布场景常用);内容来自用户自身,非不可信输入。
	goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
)

var blankLineRe = regexp.MustCompile("\n{2,}")

// AdaptContent 把用户输入的正文(sourceFormat)转换为目标平台编辑器期望的表示(targetKind)。
// sourceFormat 为空/未知时按遗留 html 行为透传,保证向后兼容。
func AdaptContent(content, sourceFormat, targetKind string) string {
	src := normalizeFormat(sourceFormat)
	// 纠偏: 内容实际是富文本 HTML(如从飞书/网页粘贴),却被标成 plain/markdown。
	// 此时若按 plain 转义会把标签变成可见文字(未渲染), 按 markdown 又会残留飞书噪声,
	// 统一改按 HTML 清洗处理。
	if (src == FormatPlain || src == FormatMarkdown) && looksLikeHTML(content) {
		src = FormatHTML
	}
	if normalizeKind(targetKind) == KindMarkdown {
		switch src {
		case FormatMarkdown:
			return content // Markdown 原生编辑器,原样注入
		case FormatPlain:
			return escapeMarkdown(content) // 防止字面符号被渲染 + 保留换行
		default:
			return cleanHTML(content) // html(如飞书粘贴): 先清洗噪声,Markdown 允许内联 HTML
		}
	}
	// 目标 = 富文本(HTML)
	switch src {
	case FormatMarkdown:
		return markdownToHTML(content)
	case FormatPlain:
		return plainToHTML(content)
	default:
		return cleanHTML(content) // html(如飞书粘贴): 清洗后再注入,避免乱码
	}
}

// htmlBlockRe 匹配富文本/网页粘贴特有的结构性标签(普通 Markdown 不会出现),
// 用来识别"被错标成 plain/markdown 的 HTML 内容"。
var htmlBlockRe = regexp.MustCompile(`(?i)<(div|p|h[1-6]|ul|ol|li|table|thead|tbody|tr|td|th|blockquote|section|article|span)[\s/>]`)

func looksLikeHTML(s string) bool {
	return htmlBlockRe.MatchString(s)
}

func normalizeFormat(f string) string {
	switch strings.ToLower(strings.TrimSpace(f)) {
	case FormatMarkdown, "md":
		return FormatMarkdown
	case FormatPlain, "text", "txt":
		return FormatPlain
	default:
		return FormatHTML
	}
}

func normalizeKind(k string) string {
	if strings.ToLower(strings.TrimSpace(k)) == KindMarkdown {
		return KindMarkdown
	}
	return KindHTML
}

func markdownToHTML(md string) string {
	var buf bytes.Buffer
	if err := mdConverter.Convert([]byte(md), &buf); err != nil {
		return plainToHTML(md) // 兜底,避免发空
	}
	return buf.String()
}

// plainToHTML 把纯文本 HTML 转义后,按空行分段为 <p>,段内单换行转 <br>。
func plainToHTML(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	var sb strings.Builder
	for _, block := range blankLineRe.Split(text, -1) {
		if strings.TrimSpace(block) == "" {
			continue
		}
		esc := html.EscapeString(strings.Trim(block, "\n"))
		esc = strings.ReplaceAll(esc, "\n", "<br>")
		sb.WriteString("<p>")
		sb.WriteString(esc)
		sb.WriteString("</p>")
	}
	return sb.String()
}

// escapeMarkdown 转义 Markdown 控制符,让纯文本在 Markdown 原生编辑器中按字面显示,
// 并以行尾两个空格保留换行(硬换行)。转义后的反斜杠会被 Markdown 渲染时消费,不影响最终展示。
func escapeMarkdown(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	const special = "\\`*_{}[]()#+-.!>~|"
	var sb strings.Builder
	for _, r := range text {
		if r == '\n' {
			sb.WriteString("  \n")
			continue
		}
		if strings.ContainsRune(special, r) {
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// keepAttrs 是清洗后元素允许保留的属性,其余(class/style/data-lark-*/data-ace-* 等噪声)一律删除。
var keepAttrs = map[string]bool{
	"href": true, "src": true, "alt": true, "colspan": true, "rowspan": true,
}

// isDropNode 判断是否整段删除该节点(飞书剪贴板元数据与脚本/样式噪声)。
func isDropNode(n *xhtml.Node) bool {
	if n.Type != xhtml.ElementNode {
		return false
	}
	switch n.DataAtom {
	case atom.Script, atom.Style, atom.Meta, atom.Link, atom.Head, atom.Title, atom.Noscript:
		return true
	}
	for _, a := range n.Attr {
		// 飞书复制残留的 30KB 元数据块,是百家号等平台乱码的主因
		if a.Key == "data-lark-record-data" {
			return true
		}
		if a.Key == "class" && strings.Contains(a.Val, "lark-record-clipboard") {
			return true
		}
	}
	return false
}

// cleanHTML 解析 HTML 片段,删除飞书剪贴板噪声节点并把元素属性精简到白名单,
// 把飞书/富文本粘贴产生的脏 HTML 还原为干净的语义 HTML,消除发布到富文本平台时的乱码。
func cleanHTML(s string) string {
	if !strings.Contains(s, "<") { // 纯文本,无需清洗
		return s
	}
	ctx := &xhtml.Node{Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body}
	nodes, err := xhtml.ParseFragment(strings.NewReader(s), ctx)
	if err != nil {
		return s // 解析失败兜底:原样返回,不阻断发布
	}
	var buf bytes.Buffer
	for _, n := range nodes {
		if isDropNode(n) {
			continue
		}
		cleanNode(n)
		_ = xhtml.Render(&buf, n)
	}
	return buf.String()
}

// cleanNode 递归删除子树中的噪声节点,并精简每个元素的属性。
func cleanNode(n *xhtml.Node) {
	var next *xhtml.Node
	for c := n.FirstChild; c != nil; c = next {
		next = c.NextSibling
		if isDropNode(c) {
			n.RemoveChild(c)
			continue
		}
		cleanNode(c)
	}
	if n.Type == xhtml.ElementNode {
		kept := n.Attr[:0]
		for _, a := range n.Attr {
			if keepAttrs[a.Key] {
				kept = append(kept, a)
			}
		}
		n.Attr = kept
	}
}
