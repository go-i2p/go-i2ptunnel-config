package i2pconv

// ConversionError represents an error during conversion
type ConversionError struct {
	Op  string
	Err error
}

// ValidationError represents an error during validation
type ValidationError struct {
	Config *TunnelConfig
	Err    error
}

func (e *ValidationError) Error() string {
	return "validation: " + e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

func (e *ConversionError) Error() string {
	return e.Op + ": " + e.Err.Error()
}

func (e *ConversionError) Unwrap() error {
	return e.Err
}
