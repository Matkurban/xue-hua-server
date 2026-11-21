package specialerror

import (
	"errors"

	"server/v3/pkg/common/errs"
)

func AddErrHandler(h func(err error) errs.CodeError) (err error) {
	if h == nil {
		return errs.New("nil handler")
	}
	errs.Handlers = append(errs.Handlers, h)
	return nil
}

func AddReplace(target error, codeErr errs.CodeError) error {
	handler := func(err error) errs.CodeError {
		if errors.Is(err, target) {
			return codeErr
		}
		return nil
	}

	if err := AddErrHandler(handler); err != nil {
		return err
	}

	return nil
}

func ErrCode(err error) errs.CodeError {
	if codeErr, ok := errors.AsType[errs.CodeError](err); ok {
		return codeErr
	}
	for i := 0; i < len(errs.Handlers); i++ {
		if codeErr := errs.Handlers[i](err); codeErr != nil {
			return codeErr
		}
	}
	return nil
}

func ErrString(err error) errs.Error {
	if codeErr, ok := errors.AsType[errs.Error](err); ok {
		return codeErr
	}
	return nil
}

func ErrWrapper(err error) errs.ErrWrapper {
	if codeErr, ok := errors.AsType[errs.ErrWrapper](err); ok {
		return codeErr
	}
	return nil
}
