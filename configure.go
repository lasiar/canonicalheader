package canonicalheader

import (
	"errors"
	"strings"
)

// stringSet is a set-of-nonempty-strings-valued flag.
type stringSet []string

func (s *stringSet) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSet) Set(flag string) error {
	list := strings.Split(flag, ",")

	*s = make(stringSet, 0, len(list))

	for _, element := range list {
		if element == "" {
			return errors.New("empty string")
		}

		*s = append(*s, element)
	}
	return nil
}
