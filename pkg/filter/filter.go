package filter

type Option interface {
	IsFilter() bool
}

type TfCheckTypeFilter struct {
	TfCheckTypes []string
}

func (t *TfCheckTypeFilter) IsFilter() bool { return true }

type DirFilter struct {
	Dir string
}

func (t *DirFilter) IsFilter() bool { return true }
