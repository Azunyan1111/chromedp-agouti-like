package chromedp_agouti_like

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"path/filepath"
	"strings"
)

type Page struct {
	CtxAlloc    context.Context
	CancelAlloc context.CancelFunc

	Ctx         context.Context
	CloseWindow context.CancelFunc
}

type Selection struct {
	Page      *Page
	Ctx       context.Context
	Query     string
	QueryType chromedp.QueryOption
}

func NewPage(isHeadless bool) (*Page, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", isHeadless))...)
	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	return &Page{CtxAlloc: allocCtx, CancelAlloc: allocCancel, Ctx: taskCtx, CloseWindow: taskCancel}, nil
}

func NewPageProxy(isHeadless bool, proxy string) (*Page, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", isHeadless), chromedp.ProxyServer(proxy))...)
	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	return &Page{CtxAlloc: allocCtx, CancelAlloc: allocCancel, Ctx: taskCtx, CloseWindow: taskCancel}, nil
}

func (p *Page) Navigate(url string) error {
	return chromedp.Run(p.Ctx, chromedp.Navigate(url))
}

func (p *Page) Find(selector string) *Selection {
	var selection Selection
	selection.Page = p
	selection.Ctx = p.Ctx
	selection.Query = selector
	selection.QueryType = chromedp.ByQuery
	return &selection
}

func (p *Page) FindXPath(xpath string) *Selection {
	var selection Selection
	selection.Page = p
	selection.Ctx = p.Ctx
	selection.Query = xpath
	selection.QueryType = chromedp.BySearch
	return &selection
}

func (s *Selection) Text() (text string, err error) {
	if s.Query == "title" {
		err = chromedp.Run(s.Ctx, chromedp.Title(&text))
		return text, err
	}
	err = chromedp.Run(s.Ctx, chromedp.Text(s.Query, &text, chromedp.NodeVisible, s.QueryType))
	return text, err
}

func (s *Selection) Value() (text string, err error) {
	err = chromedp.Run(s.Ctx, chromedp.Value(s.Query, &text, chromedp.NodeVisible, s.QueryType))
	return text, err
}

func (p Page) HTML() (html string, err error) {
	err = chromedp.Run(p.Ctx, chromedp.OuterHTML("html", &html, chromedp.NodeVisible, chromedp.ByQuery))
	return html, err
}

func (s Selection) SendKeys(key string) error {
	return chromedp.Run(s.Ctx, chromedp.SendKeys(s.Query, key, s.QueryType))
}

// document.querySelector('#container').shadowRoot.querySelector('#foo')
func (s Selection) SendKeyByShadowDOM(key string) error {
	return chromedp.Run(s.Ctx, chromedp.SendKeys(s.Query, key, chromedp.ByJSPath))
}

func (s Selection) RemoveInput() error {
	return chromedp.Run(s.Ctx, chromedp.SetValue(s.Query, "", s.QueryType))
}

func (s Selection) GetInputValue() (string, error) {
	var resp string
	err := chromedp.Run(s.Ctx, chromedp.Value(s.Query, &resp, s.QueryType))
	return resp, err
}

func (s Selection) Click() error {
	if strings.Contains(s.Query, "option") {
		// Javascript run click
		fmt.Println("document.querySelector(`" + s.Query + "`).selected = true")
		fmt.Println(s.Query)
		err := chromedp.Run(s.Ctx, Evaluate("document.querySelector(`"+s.Query+"`).selected = true;if (window.jQuery) {$(`"+s.Query+"`).change()}else{document.querySelector(`"+s.Query+"`).onchange()}"))
		_ = s.Page.Find("html").Click()
		return err
	}
	return chromedp.Run(s.Ctx, chromedp.Click(s.Query, chromedp.NodeVisible, s.QueryType))
}

func (s Selection) UploadFile(filename string) error {
	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to find absolute path for filename: %s", err)
	}
	return s.SendKeys(absFilePath)
}

func (s Selection) Attribute(attribute string) (value string, err error) {
	err = chromedp.Run(s.Ctx, chromedp.AttributeValue(s.Query, attribute, &value, nil, s.QueryType))
	return value, err
}

func (s Selection) Clear() error {
	return chromedp.Run(s.Ctx, chromedp.Clear(s.Query, s.QueryType))
}

func (s Selection) URL() (url string, err error) {
	err = chromedp.Run(s.Ctx, chromedp.Location(&url))
	return url, err
}

func (p Page) URL() (url string, err error) {
	err = chromedp.Run(p.Ctx, chromedp.Location(&url))
	return url, err
}

// TODO:Not Run
func (p Page) RunScript(script string) (resp []string, err error) {
	err = chromedp.Run(p.Ctx, chromedp.Evaluate(script, &resp))
	return resp, err
	//if strings.Contains(body,"javascript:"){
	//	return fmt.Errorf("'javascript' not support")
	//}
	//if !strings.Contains(body,"alert"){
	//	//return fmt.Errorf("alert only support")
	//}
	//
	//if arguments != nil{
	//	return fmt.Errorf("not support arguments")
	//}
	//if result != nil{
	//	return fmt.Errorf("not support result")
	//}
	//var resp string
	//err = chromedp.Run(p.Ctx,chromedp.Evaluate(body,&resp),chromedp.WaitVisible(`html`))
	//if err != nil{
	//	if err.Error() == "encountered an undefined value"{
	//		err = nil
	//		return err
	//	}
	//	if strings.Contains(err.Error(),"encountered exception"){
	//		fmt.Println("chromedp Evaluate error")
	//		fmt.Println(err.Error())
	//		return nil
	//	}
	//}
	//// Run Wait
	//_ = p.Find("html").Click()
	//return err
}

// Not Run
func Evaluate(expression string) chromedp.EvaluateAction {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// set up parameters
		p := runtime.Evaluate(expression)
		p = p.WithReturnByValue(false)

		// evaluate
		_, exp, err := p.Do(ctx)
		if err != nil {
			return err
		}
		if exp != nil {
			return exp
		}
		return nil
	})
}

// Find finds exactly one element by CSS selector.
//func (s *selectable) Find(selector string) *Selection {
//	return newSelection(s.session, s.selectors.Append(target.CSS, selector).Single())
//}
//
//// FindByXPath finds exactly one element by XPath selector.
//func (s *selectable) FindByXPath(selector string) *Selection {
//	return newSelection(s.session, s.selectors.Append(target.XPath, selector).Single())
//}
//
//// FindByLink finds exactly one anchor element by its text content.
//func (s *selectable) FindByLink(text string) *Selection {
//	return newSelection(s.session, s.selectors.Append(target.Link, text).Single())
//}
//
//// FindByLabel finds exactly one element by associated label text.
//func (s *selectable) FindByLabel(text string) *Selection {
//	return newSelection(s.session, s.selectors.Append(target.Label, text).Single())
//}
//
//// FindByButton finds exactly one button element with the provided text.
//// Supports <button>, <input type="button">, and <input type="submit">.
//func (s *selectable) FindByButton(text string) *Selection {
//	return newSelection(s.session, s.selectors.Append(target.Button, text).Single())
//}
//
//// FindByName finds exactly element with the provided name attribute.
//func (s *selectable) FindByName(name string) *Selection {
//	return newSelection(s.session, s.selectors.Append(target.Name, name).Single())
//}
