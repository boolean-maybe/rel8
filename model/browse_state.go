package model

type BrowseState struct {
	class      Class
	TableInfo  *TableInfo
	HeaderInfo *HeaderInfo
}

func (b *BrowseState) GetMode() Mode {
	return Mode{Kind: Browse, Class: b.class}
}

func (b *BrowseState) GetData() interface{} {
	return struct {
		TableInfo  *TableInfo
		HeaderInfo *HeaderInfo
	}{b.TableInfo, b.HeaderInfo}
}

func (b *BrowseState) GetAction() []Action {
	return []Action{}
}
