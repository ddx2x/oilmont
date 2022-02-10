package proc

type Function func(chan<- error)

type Proc struct {
	fs   []Function
	errC chan error
}

func NewProc() *Proc {
	proc := &Proc{
		fs:   make([]Function, 0),
		errC: make(chan error, 10),
	}
	return proc
}

func (p *Proc) Add(f ...Function) { p.fs = append(p.fs, f...) }

func (p *Proc) Start() chan error {
	for _, f := range p.fs {
		go f(p.errC)
	}
	return p.errC
}

func (p *Proc) Error() chan<- error { return p.errC }
