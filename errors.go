package env

import "errors"

var (
	// ErrorFieldIsNotWritable field is not writable.
	ErrorFieldIsNotWritable = errors.New("field is not writable")
	// ErrorFieldIsRequired field is required.
	ErrorFieldIsRequired = errors.New("field is required")
	// ErrorNotPointers source is not pointers.
	ErrorNotPointers = errors.New("must be pointers")
	// ErrorNotStruct source is not structs.
	ErrorNotStruct = errors.New("must be structs")
	// ErrorRequiredWithDefault error when required used with default value.
	ErrorRequiredWithDefault = errors.New("'required' cannot be used when a default value is specified")
	// ErrorUnrecognizedTag unrecognized tag.
	ErrorUnrecognizedTag = errors.New("unrecognized tag")
	// ErrorFieldsAreNotSupported fields are not supported.
	ErrorFieldsAreNotSupported = errors.New("fields are not supported")
	// ErrorDefaultValueForSlice default value for slice are not supported.
	ErrorDefaultValueForSlice = errors.New("default values are not supported for slice fields")
)
