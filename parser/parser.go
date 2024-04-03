package parser

type Func func(b []byte, v any) error

func (f Func) Unmarshal(b []byte, v any) error {
	return f(b, v)
}

type Parser interface {
	Unmarshal([]byte, any) error
}
