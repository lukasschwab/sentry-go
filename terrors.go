package sentry

import (
	"reflect"
	"slices"

	"github.com/lukasschwab/terrors"
)

type errorVisitor struct {
	Parent    *Exception
	Append    func(Exception)
	GetNextID func() int
}

func (v errorVisitor) Visit(err error) terrors.Visitor {
	var parentID *int
	if v.Parent != nil {
		parentID = &v.Parent.Mechanism.ExceptionID
	}

	// TODO: seems from the RFC that wrapped errors should also be considered
	// groups.
	// > When reaching one where mechanism.is_exception_group:false (or not
	// > present), include it as a "top-level" exception, and do not traverse
	// > any of its child exceptions.
	_, isExceptionGroup := err.(interface{ Unwrap() []error })

	mechanism := &Mechanism{
		IsExceptionGroup: isExceptionGroup,
		ExceptionID:      v.GetNextID(),
		ParentID:         parentID,
		// TODO: how to determine the correct mechanism type?
		Type: "generic",
	}
	exception := Exception{
		Value:      err.Error(),
		Type:       reflect.TypeOf(err).String(),
		Stacktrace: ExtractStacktrace(err),
		Mechanism:  mechanism,
	}

	v.Append(exception)

	v.Parent = &exception
	return v
}

func exceptions(err error) []Exception {
	collected := []Exception{}
	appender := func(e Exception) {
		collected = append(collected, e)
	}
	getNextID := func() int {
		return len(collected)
	}

	v := errorVisitor{Parent: nil, Append: appender, GetNextID: getNextID}
	terrors.Walk(v, err)

	slices.Reverse(collected)
	return collected
}
