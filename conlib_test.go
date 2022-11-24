package main

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewContract(t *testing.T) {
	data := ArbitraryData{
		"hello": "world",
	}
	meta := ArbitraryData{
		"hello": "meta",
	}

	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), data, meta)
	require.Equal(t, meta, contract.Meta)
	require.Equal(t, 1, len(contract.Items))

	branch := contract.Items[0]
	require.Equal(t, time.Hour*24*365, branch.Span())
	require.Equal(t, data, branch.Data)
	require.True(t, contract.Contains(branch))
}

func TestBranchLinear(t *testing.T) {
	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), ArbitraryData{"key": "world"}, nil)
	initialBranch := contract.Items[0]
	branch, err := contract.Branch(newDate(2023, 10, 10), newDate(2023, 12, 1), ArbitraryData{"key": "venus"})

	require.NoError(t, err)
	require.Equal(t, 2, len(contract.Items))
	require.Equal(t, time.Hour*24*52, branch.Span())
	require.Empty(t, initialBranch.ReplacedBy)
	require.Empty(t, branch.ReplacedBy)
	require.Equal(t, contract.Items[1], branch)
	require.Equal(t, time.Hour*24*(52+365), branch.Span()+initialBranch.Span())
	require.Equal(t, ArbitraryData{"key": "world"}, initialBranch.Data)
	require.Equal(t, ArbitraryData{"key": "venus"}, branch.Data)
}

func TestBranchLeft(t *testing.T) {
	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), ArbitraryData{"key": "world"}, nil)
	branch, err := contract.Branch(newDate(2022, 10, 10).Add(time.Hour*24*45),
		time.Time{}, ArbitraryData{"key": "venus"})

	require.NoError(t, err)

	items := contract.Items
	initial, left, head := items[0], items[1], items[2]

	require.Equal(t, 3, len(items))
	require.Equal(t, branch, head)

	require.Contains(t, initial.ReplacedBy, left.UUID)
	require.Contains(t, initial.ReplacedBy, head.UUID)

	require.Empty(t, left.ReplacedBy)
	require.Empty(t, head.ReplacedBy)

	require.Equal(t, time.Hour*24*365, initial.Span())
	require.Equal(t, time.Hour*24*45, left.Span())
	require.Equal(t, time.Hour*24*(365-45), head.Span())

	require.Equal(t, ArbitraryData{"key": "world"}, initial.Data)
	require.Equal(t, ArbitraryData{"_ref": initial.UUID}, left.Data)
	require.Equal(t, ArbitraryData{"key": "venus"}, head.Data)
}

func TestBranchRight(t *testing.T) {
	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), ArbitraryData{"key": "world"}, nil)
	branch, err := contract.Branch(newDate(2022, 10, 10),
		newDate(2022, 10, 10).Add(time.Hour*24*45), ArbitraryData{"key": "venus"})

	require.NoError(t, err)

	items := contract.Items
	initial, head, right := items[0], items[1], items[2]

	require.Equal(t, 3, len(items))
	require.Equal(t, branch, head)

	require.Contains(t, initial.ReplacedBy, right.UUID)
	require.Contains(t, initial.ReplacedBy, head.UUID)

	require.Empty(t, right.ReplacedBy)
	require.Empty(t, head.ReplacedBy)

	require.Equal(t, time.Hour*24*365, initial.Span())
	require.Equal(t, time.Hour*24*45, head.Span())
	require.Equal(t, time.Hour*24*(365-45), right.Span())

	require.Equal(t, ArbitraryData{"key": "world"}, initial.Data)
	require.Equal(t, ArbitraryData{"key": "venus"}, head.Data)
	require.Equal(t, ArbitraryData{"_ref": initial.UUID}, right.Data)
}

func TestBranchLeftRight(t *testing.T) {
	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), ArbitraryData{"key": "world"}, nil)
	branch, err := contract.Branch(newDate(2022, 10, 10).Add(time.Hour*24*5),
		newDate(2022, 10, 10).Add(time.Hour*24*45), ArbitraryData{"key": "venus"})

	require.NoError(t, err)

	items := contract.Items
	initial, left, head, right := items[0], items[1], items[2], items[3]

	require.Equal(t, 4, len(items))
	require.Equal(t, branch, head)

	require.Contains(t, initial.ReplacedBy, left.UUID)
	require.Contains(t, initial.ReplacedBy, head.UUID)
	require.Contains(t, initial.ReplacedBy, right.UUID)

	require.Empty(t, left.ReplacedBy)
	require.Empty(t, head.ReplacedBy)
	require.Empty(t, right.ReplacedBy)

	require.Equal(t, time.Hour*24*365, initial.Span())
	require.Equal(t, time.Hour*24*5, left.Span())
	require.Equal(t, time.Hour*24*40, head.Span())
	require.Equal(t, time.Hour*24*(365-45), right.Span())

	require.Equal(t, ArbitraryData{"key": "world"}, initial.Data)
	require.Equal(t, ArbitraryData{"_ref": initial.UUID}, left.Data)
	require.Equal(t, ArbitraryData{"key": "venus"}, head.Data)
	require.Equal(t, ArbitraryData{"_ref": initial.UUID}, right.Data)
}

func TestContractBranchSpansNothing(t *testing.T) {
	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), ArbitraryData{"key": "world"}, nil)

	_, err := contract.Branch(newDate(2022, 10, 10), newDate(2022, 10, 10), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "this branch would span nothing")

	_, err2 := contract.Branch(newDate(2022, 10, 10), newDate(2022, 10, 9), nil)
	require.Error(t, err2)
	require.Contains(t, err2.Error(), "this branch would span nothing")
}

func TestContractBranchSpanDateOutOfBoundary(t *testing.T) {
	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), ArbitraryData{"key": "world"}, nil)

	_, err := contract.Branch(newDate(2023, 10, 10).Add(time.Hour*24*52), time.Time{}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "out of the boundary")

	_, err2 := contract.Branch(newDate(2022, 10, 10).Truncate(time.Hour*24*52), time.Time{}, nil)
	require.NoError(t, err2)

	items := contract.Items
	head, head1 := items[0], items[1]
	require.Contains(t, head.ReplacedBy, head1.UUID)
	require.True(t, head1.StartAt.Before(head.StartAt))
}

func TestBranchCaseNoOverlap(t *testing.T) {
	contract := NewContract(newDate(2022, 10, 10), newDate(2023, 10, 10), ArbitraryData{"key": "world"}, nil)
	_, _ = contract.Branch(newDate(2022, 10, 11), time.Time{}, nil)
	_, _ = contract.Branch(newDate(2022, 10, 13), time.Time{}, nil)

	items := contract.Items
	_, left, head, left2, head2 := items[0], items[1], items[2], items[3], items[4]
	require.Equal(t, 5, len(items))

	require.Equal(t, time.Hour*24*364, head.Span())
	require.Equal(t, time.Hour*24*1, left.Span())
	require.Equal(t, time.Hour*24*362, head2.Span())
	require.Equal(t, time.Hour*24*2, left2.Span())
}

func TestContract_ResolveDataref(t *testing.T) {
	startAt, endAt := newDate(2022, 10, 10), newDate(2023, 10, 10)

	contract := NewContract(startAt, endAt, ArbitraryData{"key": "world"}, nil)
	_, _ = contract.Branch(startAt.Add(time.Hour*24*30), startAt.Add(time.Hour*24*60), ArbitraryData{"key": "venus"})
	_, _ = contract.Branch(startAt.Add(time.Hour*24*35), startAt.Add(time.Hour*24*40), ArbitraryData{"key": "jupiter"})
	_, _ = contract.Branch(startAt.Add(time.Hour*24*60), time.Time{}, ArbitraryData{"key": "mars"})

	items := contract.Items
	m0, l1, m1, r1, l2, m2, r2, m3 := items[0], items[1], items[2], items[3], items[4], items[5], items[6], items[7]

	require.Equal(t, ArbitraryData{"key": "world"}, m0.Data)

	require.Equal(t, ArbitraryData{"_ref": m0.UUID}, l1.Data)
	require.Equal(t, ArbitraryData{"key": "venus"}, m1.Data)
	require.Equal(t, ArbitraryData{"_ref": m0.UUID}, r1.Data)

	require.Equal(t, ArbitraryData{"_ref": m1.UUID}, l2.Data)
	require.Equal(t, ArbitraryData{"key": "jupiter"}, m2.Data)
	require.Equal(t, ArbitraryData{"_ref": m1.UUID}, r2.Data)

	require.Equal(t, ArbitraryData{"key": "mars"}, m3.Data)
}
