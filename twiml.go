// Package twiml provides Twilio Markup Language support for building web
// services with instructions for twilio how to handle incoming call or message.
package twiml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

// Markup interface is satisfied by valid TwiML verbs.
type Markup interface {
	// Type returns the TwiML verb name for use in pattern-matching
	Type() string
	// Validate will verify that TwiML responses are properly constructed of allowed options for all fields
	Validate() error
}

// ValidationError will return one or more errors encountered during validation
type ValidationError struct {
	Errors []error
}

// Error() returns a custom string representation of all errors encountered during validation
func (v ValidationError) Error() string {
	e := []string{"Invalid TwiML markup:"}
	for _, err := range v.Errors {
		e = append(e, err.Error())
	}
	return strings.Join(e, "\n")
}

// Response container for other TwiML verbs
type Response struct {
	XMLName                xml.Name `xml:"Response"`
	IgnoreValidationErrors bool     `xml:"-"`
	Children               []Markup
}

// Type returns the XML name of the verb
func (r *Response) Type() string {
	return "Response"
}

// Validate recursively validates all nested verbs within the response, returning a ValidationError
// if any are constructed improperly
func (r *Response) Validate() error {
	if len(r.Children) == 0 {
		return ValidationError{[]error{fmt.Errorf("Can not encode an empty response")}}
	}
	var errs []error
	for _, s := range r.Children {
		switch t := s.Type(); t {
		case "Enqueue", "Hangup", "Leave", "Pause", "Play", "Record", "Redirect", "Reject", "Say", "Dial", "Gather", "Sms", "Start", "Stop", "Connect":
			if childErr := s.Validate(); childErr != nil {
				errs = append(errs, childErr)
			}
		default:
			return ValidationError{[]error{fmt.Errorf("Unknown markup type %T as child of Response", s)}}
		}
	}
	if len(errs) > 0 {
		return ValidationError{errs}
	}
	return nil
}

// NewResponse creates new response container.  Use Add() to chain together the response from allowed verbs.
func NewResponse() *Response {
	resp := new(Response)
	return resp
}

// Add appends TwiML verb structs to response. Valid verbs: Enqueue, Say,
// Leave, Message, Pause, Play, Record, Redirect, Reject, Hangup
func (r *Response) Add(ml ...Markup) {
	for _, s := range ml {
		r.Children = append(r.Children, s)
	}
	return
}

// Encode returns an XML encoded response or a ValidationError if any
// markup fails validation.
func (r *Response) Encode() ([]byte, error) {
	var buf = new(bytes.Buffer)

	if err := r.Validate(); err != nil {
		return buf.Bytes(), err
	}

	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")

	_, err := buf.Write([]byte(xml.Header))
	if err != nil {
		return buf.Bytes(), err
	}

	if err := enc.Encode(r); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

// String returns a formatted XML response
func (r *Response) String() (string, error) {
	b, err := r.Encode()
	return string(b), err
}
