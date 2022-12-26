package expenses

const (
	Endpoint = "expenses"

	CreateBody        = `{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]}`
	InvalidCreateBody = `{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]`

	UpdateBody        = `{"id":1,"title":"test expense update","amount":200,"note":"test note update","tags":["tag1","tag2"]}`
	InvalidUpdateBody = `{"title":"test expense","amount":100,"note":"test note","tags":["tag1","tag2"]`

	Token        = "January 2, 2006"
	InvalidToken = "invalid token"
)
