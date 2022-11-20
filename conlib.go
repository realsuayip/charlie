package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"time"
)

type Branch struct {
	UUID                      string
	Data                      ArbitraryData
	ReplacedBy                []string // List of UUID strings
	StartAt, EndAt, CreatedAt time.Time
}

func NewBranch(StartAt, EndAt time.Time, Data ArbitraryData) *Branch {
	return &Branch{
		UUID:      uuid.New().String(),
		Data:      Data,
		StartAt:   StartAt.Truncate(time.Second),
		EndAt:     EndAt.Truncate(time.Second),
		CreatedAt: time.Now().UTC(),
	}
}

func (b *Branch) Span() time.Duration {
	return b.EndAt.Sub(b.StartAt)
}

func (b *Branch) String() string {
	var replacedBy []string
	for _, item := range b.ReplacedBy {
		replacedBy = append(replacedBy, item[0:8])
	}
	return fmt.Sprintf(
		"<Branch %-10s StartAt=%-22s EndAt=%-22s Span=%-14s Data=%s ReplacedBy=%s>",
		b.UUID[0:8],
		b.StartAt.Format(time.RFC3339),
		b.EndAt.Format(time.RFC3339),
		duration(b.Span()),
		b.Data,
		replacedBy)
}

type Contract struct {
	UUID      string
	Meta      ArbitraryData
	Items     []*Branch
	CreatedAt time.Time
}

func NewContract(StartAt, EndAt time.Time, Data ArbitraryData, Meta ArbitraryData) *Contract {
	return &Contract{
		UUID:      uuid.New().String(),
		Meta:      Meta,
		Items:     []*Branch{NewBranch(StartAt, EndAt, Data)},
		CreatedAt: time.Now().UTC(),
	}
}

func (c *Contract) Contains(branch *Branch) bool {
	for _, item := range c.Items {
		if item.UUID == branch.UUID {
			return true
		}
	}
	return false
}

func (c *Contract) Shift(old, new *Branch, replace bool) {
	if !c.Contains(new) {
		c.Items = append(c.Items, new)
	}

	if replace {
		old.ReplacedBy = append(old.ReplacedBy, new.UUID)
	}
}

func (c *Contract) Branch(StartAt, EndAt time.Time, Data ArbitraryData) (*Branch, error) {
	items := lo.Filter[*Branch](c.Items, func(b *Branch, index int) bool {
		return len(b.ReplacedBy) == 0
	})

	timeMap := lo.Map[*Branch, time.Time]
	ends := timeMap(items, func(b *Branch, index int) time.Time { return b.EndAt })
	starts := timeMap(items, func(b *Branch, index int) time.Time { return b.StartAt })

	maxEnd := lo.MaxBy(ends, func(t1, t2 time.Time) bool { return t1.After(t2) })
	minStart := lo.MinBy(starts, func(t1, t2 time.Time) bool { return t1.Before(t2) })

	if StartAt.After(maxEnd) || StartAt.Before(minStart) {
		return nil, fmt.Errorf(
			"given start date (%s) is out of the boundary, the valid boundary is between %s and %s (inclusively)",
			StartAt, minStart, maxEnd)
	}

	if time.Time.IsZero(EndAt) {
		EndAt = maxEnd
	}

	branch := NewBranch(StartAt, EndAt, Data)

	if branch.Span() <= 0 {
		return branch, fmt.Errorf("this branch would span nothing (%s)", branch)
	}

	for _, item := range items {
		latestStart := maxDate(item.StartAt, StartAt)
		earliestEnd := minDate(item.EndAt, EndAt)

		delta := earliestEnd.Sub(latestStart)
		overlap := maxDuration(0, delta)
		isExtension := (overlap == 0) && (StartAt == item.EndAt)

		if !(overlap != 0 || isExtension) {
			continue
		}

		ldelta := maxDuration(0, StartAt.Sub(item.StartAt))
		rdelta := maxDuration(0, item.EndAt.Sub(EndAt))
		// dataref := make(map[string]interface{}) todo

		if (ldelta != 0) && (ldelta != item.Span()) {
			left := NewBranch(item.StartAt, item.StartAt.Add(ldelta), item.Data)
			c.Shift(item, left, true)
		}

		c.Shift(item, branch, !isExtension)

		if rdelta != 0 {
			right := NewBranch(EndAt, item.EndAt, item.Data)
			c.Shift(item, right, true)
		}
	}
	return branch, nil
}

func (c *Contract) Explain() {
	var span time.Duration = 0
	for _, item := range c.Items {
		prefix := "*"

		if len(item.ReplacedBy) == 0 {
			span += item.Span()
			prefix = ";"
		}
		fmt.Print(fmt.Sprintf("%s %s\n", prefix, item))
	}
	fmt.Println(fmt.Sprintf("span=%s,count=%d", duration(span), len(c.Items)))
}
