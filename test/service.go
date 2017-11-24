package test

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
	Get(a, b []*string, _ error) (int, int) // inline comment
	Post()
}

type StringServices interface {
	Get(a, b []*string, _ error) (i, j int, writer io.Writer) // inline comment
}

type OMG struct {
	i string `json:"i,j"xml:"i,k"gorm:"dasdas,f"`
	j []int
}

func yy(x ...[]*[]*map[*[]interface {
	Null()
}]map[*int][][]****int64) {

}
