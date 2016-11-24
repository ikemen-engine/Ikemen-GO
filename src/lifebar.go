package main

type Lifebar struct{ aniTbl *AnimationTable }

func LoadLifebar(deffile string) (*Lifebar, error) {
	l := &Lifebar{aniTbl: NewAnimationTable()}
	str, err := LoadText(deffile)
	if err != nil {
		return nil, err
	}
	lines := SplitAndTrim(str, "\n")
	return l, nil
}
