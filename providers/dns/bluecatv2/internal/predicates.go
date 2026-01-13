package internal

import (
	"fmt"
	"strings"
)

type Predicate struct {
	field    string
	operator string
	values   []string
}

func (p *Predicate) String() string {
	var values []string
	for _, v := range p.values {
		values = append(values, fmt.Sprintf("'%s'", v))
	}

	return fmt.Sprintf("%s:%s(%s)", p.field, p.operator, strings.Join(values, ", "))
}

func Eq(field, value string) *Predicate {
	return &Predicate{field: field, operator: "eq", values: []string{value}}
}

func Contains(field, value string) *Predicate {
	return &Predicate{field: field, operator: "contains", values: []string{value}}
}

func StartsWith(field, value string) *Predicate {
	return &Predicate{field: field, operator: "startsWith", values: []string{value}}
}

func EndsWith(field, value string) *Predicate {
	return &Predicate{field: field, operator: "endsWith", values: []string{value}}
}

func In(field string, values ...string) *Predicate {
	return &Predicate{field: field, operator: "in", values: values}
}

type Combined struct {
	predicates []*Predicate
	operator   string
}

func (o *Combined) String() string {
	var parts []string

	for _, predicate := range o.predicates {
		parts = append(parts, predicate.String())
	}

	return strings.Join(parts, " "+o.operator+" ")
}

func And(predicates ...*Predicate) *Combined {
	return &Combined{predicates: predicates, operator: "and"}
}

func Or(predicates ...*Predicate) *Combined {
	return &Combined{predicates: predicates, operator: "or"}
}
