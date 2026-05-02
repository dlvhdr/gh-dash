package inputbox

// TODO: use Inserter and remove it from Source
type Inserter interface {
	Insert(input string, insert string) string
}
