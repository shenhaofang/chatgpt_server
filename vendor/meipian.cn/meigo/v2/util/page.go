package util

type Page struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
}

func (p *Page) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p *Page) GetPage() int {
	return p.Page
}

func (p *Page) GetPagesize() int {
	return p.PageSize
}

// 初始化默认参数
func (p *Page) Init() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 20
	}
}