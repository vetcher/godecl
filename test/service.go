/*
here are the docs
f
sadas
*/
// wtf
package stringsvc

// import comment
import (
	"io"
	"strings"
)

var (
	x, y = 5, strings.Reader{}
	z    strings.Reader
	u    = x
	m    map[string]io.Writer
)

// comment
const l = "l"

// interface docs
type StringService interface {
	// inside interface docs

	// Get docs
	Get(a, b []*string, _ error) (int, int) // inline comment
}

//hfsdfjsd
type StringServices interface {
	// inside interface docs

	// Get docs
	Get(a, b []*string, _ error) (i, j int, writer io.Writer) // inline comment
}

type OMG struct {
	i string
}
