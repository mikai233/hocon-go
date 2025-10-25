package merge

import "hocon-go/common"

type Memo struct {
	Tracker            []common.Path
	SubstituionCounter uint
}
