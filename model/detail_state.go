package model

type DetailState struct {
	class Class
	text  string
}

func (b *DetailState) GetMode() Mode {
	return Mode{Kind: Detail, Class: b.class}
}

func (b *DetailState) GetData() interface{} {
	return struct {
		text string
	}{b.text}
}

func (b *DetailState) GetAction() []Action {
	return []Action{}
}
