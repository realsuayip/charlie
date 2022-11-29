package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Branch struct {
	ID         primitive.ObjectID   `bson:"_id" json:"_id"`
	Data       ArbitraryData        `bson:"data" json:"data"`
	ReplacedBy []primitive.ObjectID `bson:"replaced_by" json:"replaced_by"`
	StartAt    time.Time            `bson:"start_at" json:"start_at"`
	EndAt      time.Time            `bson:"end_at" json:"end_at"`
	CreatedAt  time.Time            `bson:"created_at" json:"created_at"`
}

func NewBranch(StartAt, EndAt time.Time, Data ArbitraryData) *Branch {
	return &Branch{
		ID:        primitive.NewObjectID(),
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
	var replacedBy []primitive.ObjectID
	for _, item := range b.ReplacedBy {
		replacedBy = append(replacedBy, item)
	}
	return fmt.Sprintf(
		"<Branch %-10s StartAt=%-22s EndAt=%-22s Span=%-14s Data=%s ReplacedBy=%s>",
		b.ID,
		b.StartAt.Format(time.RFC3339),
		b.EndAt.Format(time.RFC3339),
		duration(b.Span()),
		b.Data,
		replacedBy)
}

type Contract struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	Meta      ArbitraryData      `bson:"meta" json:"meta"`
	Items     []*Branch          `bson:"items" json:"items"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func NewContract(StartAt, EndAt time.Time, Data ArbitraryData, Meta ArbitraryData) *Contract {
	return &Contract{
		ID:        primitive.NewObjectID(),
		Meta:      Meta,
		Items:     []*Branch{NewBranch(StartAt, EndAt, Data)},
		CreatedAt: time.Now().UTC(),
	}
}

func (c *Contract) Contains(branch *Branch) bool {
	for _, item := range c.Items {
		if item.ID == branch.ID {
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
		old.ReplacedBy = append(old.ReplacedBy, new.ID)
	}
}

func (c *Contract) ResolveDataref(b *Branch) ArbitraryData {
	if ref, exists := b.Data["_ref"]; exists {
		for _, branch := range c.Items {
			if ref == branch.ID {
				return c.ResolveDataref(branch)
			}
		}
	}
	return ArbitraryData{
		"_ref": b.ID,
	}
}

func (c *Contract) Branch(StartAt, EndAt time.Time, Data ArbitraryData) (*Branch, error) {
	var minStart, maxEnd time.Time
	var items []*Branch

	for _, item := range c.Items {
		if len(item.ReplacedBy) == 0 {
			items = append(items, item)

			if item.EndAt.After(maxEnd) {
				maxEnd = item.EndAt
			}
		}
	}

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

		lsplit := (ldelta != 0) && (ldelta != item.Span())
		rsplit := rdelta != 0

		var dataref ArbitraryData
		if rsplit || lsplit {
			dataref = c.ResolveDataref(item)
		}

		if lsplit {
			left := NewBranch(item.StartAt, item.StartAt.Add(ldelta), dataref)
			c.Shift(item, left, true)
		}

		c.Shift(item, branch, !isExtension)

		if rsplit {
			right := NewBranch(EndAt, item.EndAt, dataref)
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
