package extract

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	mdrender "research-toolkit/lib/md-render"
	"strconv"
	"strings"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type Extractor struct {
	ctx    context.Context
	cancel func()
}

func NewExtractor() (Extractor, error) {
	dataTemp := "chrome-data"
	err := os.RemoveAll(dataTemp)
	if err != nil {
		return Extractor{}, err
	}
	err = os.Mkdir(dataTemp, 0777)
	if err != nil {
		return Extractor{}, err
	}

	allocatorCtx, _ := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ExecPath("/usr/bin/thorium-browser"),
			chromedp.UserDataDir(dataTemp),
			chromedp.Flag("load-extension", "ublock"),
			chromedp.Flag("headless", false),
			chromedp.Flag("blink-settings", "imagesEnabled=false"),
			chromedp.Flag("disable-extensions", false),
		)...,
	)
	ctx, cancel := chromedp.NewContext(allocatorCtx)
	if err != nil {
		return Extractor{}, err
	}

	err = chromedp.Run(ctx)
	if err != nil {
		return Extractor{}, err
	}

	return Extractor{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (e Extractor) Extract(url *url.URL) ([]mdrender.Node, AXNode, error) {
	currentCtx, cancel := chromedp.NewContext(e.ctx)
	defer cancel()

	axTree := AXNode{}
	var mdTree []mdrender.Node
	err := chromedp.Run(
		currentCtx,
		chromedp.Navigate(url.String()),
		chromedp.ActionFunc(func(pageCtx context.Context) error {
			err := accessibility.Enable().Do(pageCtx)
			if err != nil {
				return err
			}

			axTree, err = getAccessibilityTree(pageCtx)
			if err != nil {
				return err
			}

			mdTree = MarkdownFromAXTree(pageCtx, axTree)

			return nil
		}),
	)

	return mdTree, axTree, err
}

func (e Extractor) Context() context.Context {
	return e.ctx
}

func (e Extractor) Destroy() {
	e.cancel()
}

func filterParagraphElements(nodes []mdrender.Node) []mdrender.ParagraphElement {
	result := []mdrender.ParagraphElement{}
	for _, n := range nodes {
		cast, ok := n.(mdrender.ParagraphElement)
		if ok {
			result = append(result, cast)
		}
	}
	return result
}

type traversalState struct {
	underMain      bool
	underParagraph bool
	listItemDepth  int
}

func MarkdownFromAXTree(pageCtx context.Context, root AXNode) []mdrender.Node {
	return convertMdFromAx(pageCtx, root, traversalState{})
}

func convertMdFromAx(
	pageCtx context.Context,
	root AXNode,
	state traversalState,
) []mdrender.Node {
	childState := state
	nodes := []mdrender.Node{}

	switch string(root.Role) {
	case "heading":
		children := []mdrender.Node{}
		for _, child := range root.Children {
			children = append(
				children,
				convertMdFromAx(pageCtx, child, state)...,
			)
		}

		order := 1
		node, err := dom.DescribeNode().
			WithBackendNodeID(cdp.BackendNodeID(root.DomNodeId)).
			Do(pageCtx)
		if err != nil {
			slog.Warn("could not get DOM node", "id", root.DomNodeId, "err", err.Error())
		} else {
			parsed, err := strconv.ParseInt(strings.ReplaceAll(node.NodeName, "H", ""), 10, 32)
			if err != nil {
				slog.Warn("could not parse heading order", "tagName", node.NodeName)
			} else {
				order = int(parsed)
			}
		}

		elements := filterParagraphElements(children)
		if len(elements) == 0 {
			return []mdrender.Node{}
		}

		return []mdrender.Node{
			mdrender.Header{
				Order: order,
				Content: mdrender.Paragraph{
					Elements: elements,
				},
			},
		}
	case "list":
		listType := mdrender.LIST_UNORDERED
		node, err := dom.DescribeNode().
			WithBackendNodeID(cdp.BackendNodeID(root.DomNodeId)).
			Do(pageCtx)
		if err != nil {
			slog.Warn("could not get DOM node", "id", root.DomNodeId, "err", err.Error())
		} else if node.Name == "OL" {
			listType = mdrender.LIST_ORDERED
		}

		childState.listItemDepth++

		children := []mdrender.ListItem{}
		for _, child := range root.Children {
			childNodes := convertMdFromAx(pageCtx, child, childState)
			for _, childNode := range childNodes {
				cast, ok := childNode.(mdrender.ListItem)
				if ok {
					children = append(children, cast)
				}
			}
		}
		if len(children) == 0 {
			return []mdrender.Node{}
		}

		return []mdrender.Node{
			mdrender.List{
				Type:  listType,
				Items: children,
			},
		}
	case "paragraph", "note":
		if state.underParagraph {
			children := []mdrender.Node{}
			for _, child := range root.Children {
				children = append(children, convertMdFromAx(pageCtx, child, state)...)
			}
			return children
		}

		childState.underParagraph = true

		children := []mdrender.ParagraphElement{}
		for _, child := range root.Children {
			children = append(
				children,
				filterParagraphElements(convertMdFromAx(
					pageCtx, child, childState,
				))...,
			)
		}

		if len(children) == 0 {
			return []mdrender.Node{}
		}

		return []mdrender.Node{
			mdrender.Paragraph{
				Elements: children,
			},
		}
	case "code":
		if len(root.Children) == 0 {
			return []mdrender.Node{}
		}

		child := root.Children[0]
		if child.Role != "StaticText" {
			return []mdrender.Node{}
		}

		return []mdrender.Node{
			mdrender.BlockCode{
				Content: child.Name,
			},
		}
	case "link":
		children := []mdrender.Node{}
		for _, child := range root.Children {
			children = append(
				children,
				convertMdFromAx(pageCtx, child, state)...,
			)
		}

		href := "ERROR"
		node, err := dom.DescribeNode().
			WithBackendNodeID(cdp.BackendNodeID(root.DomNodeId)).
			Do(pageCtx)
		if err != nil {
			slog.Warn("could not get DOM node", "id", root.DomNodeId, "err", err.Error())
		} else {
			href = node.AttributeValue("href")
		}

		elements := filterParagraphElements(children)
		if len(elements) == 0 {
			return []mdrender.Node{}
		}

		return []mdrender.Node{
			mdrender.Link{
				URL: href,
				Title: mdrender.Paragraph{
					Elements: elements,
				},
			},
		}
	case "StaticText":
		text := strings.Trim(root.Name, " \t\n")
		if text == "" {
			return []mdrender.Node{}
		}
		return []mdrender.Node{
			mdrender.PlainText{
				Content: text,
			},
		}
	case "RootWebArea":
		header := strings.Trim(root.Name, " \t\n")
		if header != "" {
			nodes = append(nodes, mdrender.Header{
				Order: 1,
				Content: mdrender.PlainText{
					Content: header,
				},
			})
		}
	case "main":
		childState.underMain = true
	}

	for _, child := range root.Children {
		nodes = append(
			nodes,
			convertMdFromAx(pageCtx, child, childState)...,
		)
	}

	return nodes
}
