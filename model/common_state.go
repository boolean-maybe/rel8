package model

type HeaderInfo map[string]string

type CommonState struct {
	HeaderInfo *HeaderInfo
}

func (cs *CommonState) GetHeaderInfo() *HeaderInfo {
	return cs.HeaderInfo
}
