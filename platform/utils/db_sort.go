package utils

import (
	"fmt"
	"strings"

	"github.com/tucanbit/internal/constant"
	"github.com/tucanbit/internal/constant/errors"
)

func ValidateSortOptions(name, req string) error {
	if strings.ToLower(req) != constant.SORT_QUERY_ASC && strings.ToLower(req) != constant.SORT_QUERY_DESC {
		err := fmt.Errorf("invalid  sort %s options is given only acs or desc allowed", name)
		err = errors.ErrInvalidUserInput.Wrap(err, err.Error())
		return err
	}
	return nil
}
