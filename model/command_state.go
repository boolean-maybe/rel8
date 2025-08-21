package model

type CommandState struct {
	class Class
	text  string
}

func (b *CommandState) GetMode() Mode {
	return Mode{Kind: Command, Class: b.class}
}

func (b *CommandState) GetData() interface{} {
	return struct {
		text string
	}{b.text}
}

func (b *CommandState) GetAction() []Action {
	return []Action{}
}
