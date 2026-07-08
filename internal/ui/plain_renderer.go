package ui

import "fmt"

type plainRenderer struct{}

type plainSpinnerHandle struct {
	spinner *Spinner
}

func (plainRenderer) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s✓ %s%s\n", ColorGreen, msg, ColorReset)
}

func (plainRenderer) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s✗ %s%s\n", ColorRed, msg, ColorReset)
}

func (plainRenderer) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s⚠ %s%s\n", ColorYellow, msg, ColorReset)
}

func (plainRenderer) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%sℹ %s%s\n", ColorBlue, msg, ColorReset)
}

func (plainRenderer) NewSpinner(message string) SpinnerHandle {
	s := NewSpinner(message)
	s.Start()
	return &plainSpinnerHandle{spinner: s}
}

func (h *plainSpinnerHandle) Stop() {
	if h.spinner != nil {
		h.spinner.Stop()
	}
}

func (h *plainSpinnerHandle) UpdateMessage(message string) {
	if h.spinner != nil {
		h.spinner.UpdateMessage(message)
	}
}
