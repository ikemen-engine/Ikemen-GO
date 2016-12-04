package main

type CharGlobalInfo struct {
	def              string
	sff              *Sff
	palno, drawpalno int32
	wakewakaLength   int
}
type Char struct {
	key         int
	helperindex int
	playerno    int
	keyctrl     bool
	player      bool
}

func newChar(n, idx int) (c *Char) {
	c = &Char{}
	c.init(n, idx)
	return c
}
func (c *Char) init(n, idx int) {
	c.playerno, c.helperindex = n, idx
	if c.helperindex == 0 {
		c.keyctrl, c.player = true, true
	}
	c.key = n
	if n >= 0 && n < len(sys.com) && sys.com[n] != 0 {
		c.key ^= -1
	}
}
func (c *Char) load(def string) error {
	unimplemented()
	return nil
}
