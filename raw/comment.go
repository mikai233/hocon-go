package raw

// ---------------- CommentType ----------------

type CommentType int

const (
	DoubleSlash CommentType = iota
	Hash
)

func (c CommentType) String() string {
	switch c {
	case DoubleSlash:
		return "//"
	case Hash:
		return "#"
	default:
		return ""
	}
}

type Comment struct {
	Content string
	Ty      CommentType
}

func NewComment(content string, ty CommentType) *Comment {
	return &Comment{Content: content, Ty: ty}
}

func NewDoubleSlashComment(content string) *Comment {
	return NewComment(content, DoubleSlash)
}

func NewHashComment(content string) *Comment {
	return NewComment(content, Hash)
}

func (c *Comment) String() string {
	return c.Ty.String() + c.Content
}

func CommentFromStr(val string) *Comment {
	return NewDoubleSlashComment(val)
}

func CommentFromString(val string) *Comment {
	return NewHashComment(val)
}
