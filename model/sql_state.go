package model

type SqlState struct {
	class Class
	sql   string
}

func (b *SqlState) GetMode() Mode {
	return Mode{Kind: SQL, Class: b.class}
}

func (b *SqlState) GetData() interface{} {
	return struct {
		sql string
	}{b.sql}
}

func (b *SqlState) GetAction() []Action {
	return []Action{}
}

//todo where is database type coming
//todo when is Sql or other info updated - not on each key
//todo there can be lots of Details views e.g. Record, Table description
//todo DetailState - ??? type of detail? and text of Detail?
//todo depending on type (e.g. table SQL) - different hot keys
//todo actions are per each smallest state variation?

//todo each component includes a Box that has SetInputCapture

//todo App.SetInputCapture sets a function which captures all key events before they are forwarded to the key event handler
// of the primitive which currently has focus. This function can then choose to forward that key event (or a different one)
// by returning it or stop the key event processing by returning nil.

//todo in k9s ui.Pages IS tview.Pages AND IS a stack

//todo view has current state from state manager to decide when and what to send
//todo collect all necessary info (selection text etc.) into event and invoke handler
