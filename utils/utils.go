package utils

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cockroachdb/errors"
)

func Reader(a any) (io.Reader, error) {
	if a == nil {
		return nil, errors.New("nil value when converting to io.Reader")
	}
	switch obj := a.(type) {
	case io.Reader:
		return obj, nil
	case []byte:
		return bytes.NewBuffer(obj), nil
	case string:
		return bytes.NewBufferString(obj), nil
	case fmt.Stringer:
		return bytes.NewBufferString(obj.String()), nil
	default:
		return nil, errors.Newf("unsupported type %T when converting to io.Reader", a)
	}
}
