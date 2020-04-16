// Code generated by genny. DO NOT EDIT.
// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/mauricelam/genny

package set

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stackrox/rox/generated/storage"
)

// If you want to add a set for your custom type, simply add another go generate line along with the
// existing ones. If you're creating a set for a primitive type, you can follow the example of "string"
// and create the generated file in this package.
// For non-primitive sets, please make the generated code files go outside this package.
// Sometimes, you might need to create it in the same package where it is defined to avoid import cycles.
// The permission set is an example of how to do that.
// You can also specify the -imp command to specify additional imports in your generated file, if required.

// storage.UpgradeProgress_UpgradeState represents a generic type that we want to have a set of.

// StorageUpgradeProgress_UpgradeStateSet will get translated to generic sets.
type StorageUpgradeProgress_UpgradeStateSet map[storage.UpgradeProgress_UpgradeState]struct{}

// Add adds an element of type storage.UpgradeProgress_UpgradeState.
func (k *StorageUpgradeProgress_UpgradeStateSet) Add(i storage.UpgradeProgress_UpgradeState) bool {
	if *k == nil {
		*k = make(map[storage.UpgradeProgress_UpgradeState]struct{})
	}

	oldLen := len(*k)
	(*k)[i] = struct{}{}
	return len(*k) > oldLen
}

// AddMatching is a utility function that adds all the elements that match the given function to the set.
func (k *StorageUpgradeProgress_UpgradeStateSet) AddMatching(matchFunc func(storage.UpgradeProgress_UpgradeState) bool, elems ...storage.UpgradeProgress_UpgradeState) bool {
	oldLen := len(*k)
	for _, elem := range elems {
		if !matchFunc(elem) {
			continue
		}
		if *k == nil {
			*k = make(map[storage.UpgradeProgress_UpgradeState]struct{})
		}
		(*k)[elem] = struct{}{}
	}
	return len(*k) > oldLen
}

// AddAll adds all elements of type storage.UpgradeProgress_UpgradeState. The return value is true if any new element
// was added.
func (k *StorageUpgradeProgress_UpgradeStateSet) AddAll(is ...storage.UpgradeProgress_UpgradeState) bool {
	if len(is) == 0 {
		return false
	}
	if *k == nil {
		*k = make(map[storage.UpgradeProgress_UpgradeState]struct{})
	}

	oldLen := len(*k)
	for _, i := range is {
		(*k)[i] = struct{}{}
	}
	return len(*k) > oldLen
}

// Remove removes an element of type storage.UpgradeProgress_UpgradeState.
func (k *StorageUpgradeProgress_UpgradeStateSet) Remove(i storage.UpgradeProgress_UpgradeState) bool {
	if len(*k) == 0 {
		return false
	}

	oldLen := len(*k)
	delete(*k, i)
	return len(*k) < oldLen
}

// RemoveAll removes the given elements.
func (k *StorageUpgradeProgress_UpgradeStateSet) RemoveAll(is ...storage.UpgradeProgress_UpgradeState) bool {
	if len(*k) == 0 {
		return false
	}

	oldLen := len(*k)
	for _, i := range is {
		delete(*k, i)
	}
	return len(*k) < oldLen
}

// RemoveMatching removes all elements that match a given predicate.
func (k *StorageUpgradeProgress_UpgradeStateSet) RemoveMatching(pred func(storage.UpgradeProgress_UpgradeState) bool) bool {
	if len(*k) == 0 {
		return false
	}

	oldLen := len(*k)
	for elem := range *k {
		if pred(elem) {
			delete(*k, elem)
		}
	}
	return len(*k) < oldLen
}

// Contains returns whether the set contains an element of type storage.UpgradeProgress_UpgradeState.
func (k StorageUpgradeProgress_UpgradeStateSet) Contains(i storage.UpgradeProgress_UpgradeState) bool {
	_, ok := k[i]
	return ok
}

// Cardinality returns the number of elements in the set.
func (k StorageUpgradeProgress_UpgradeStateSet) Cardinality() int {
	return len(k)
}

// IsEmpty returns whether the underlying set is empty (includes uninitialized).
func (k StorageUpgradeProgress_UpgradeStateSet) IsEmpty() bool {
	return len(k) == 0
}

// Clone returns a copy of this set.
func (k StorageUpgradeProgress_UpgradeStateSet) Clone() StorageUpgradeProgress_UpgradeStateSet {
	if k == nil {
		return nil
	}
	cloned := make(map[storage.UpgradeProgress_UpgradeState]struct{}, len(k))
	for elem := range k {
		cloned[elem] = struct{}{}
	}
	return cloned
}

// Difference returns a new set with all elements of k not in other.
func (k StorageUpgradeProgress_UpgradeStateSet) Difference(other StorageUpgradeProgress_UpgradeStateSet) StorageUpgradeProgress_UpgradeStateSet {
	if len(k) == 0 || len(other) == 0 {
		return k.Clone()
	}

	retained := make(map[storage.UpgradeProgress_UpgradeState]struct{}, len(k))
	for elem := range k {
		if !other.Contains(elem) {
			retained[elem] = struct{}{}
		}
	}
	return retained
}

// Intersect returns a new set with the intersection of the members of both sets.
func (k StorageUpgradeProgress_UpgradeStateSet) Intersect(other StorageUpgradeProgress_UpgradeStateSet) StorageUpgradeProgress_UpgradeStateSet {
	maxIntLen := len(k)
	smaller, larger := k, other
	if l := len(other); l < maxIntLen {
		maxIntLen = l
		smaller, larger = larger, smaller
	}
	if maxIntLen == 0 {
		return nil
	}

	retained := make(map[storage.UpgradeProgress_UpgradeState]struct{}, maxIntLen)
	for elem := range smaller {
		if _, ok := larger[elem]; ok {
			retained[elem] = struct{}{}
		}
	}
	return retained
}

// Union returns a new set with the union of the members of both sets.
func (k StorageUpgradeProgress_UpgradeStateSet) Union(other StorageUpgradeProgress_UpgradeStateSet) StorageUpgradeProgress_UpgradeStateSet {
	if len(k) == 0 {
		return other.Clone()
	} else if len(other) == 0 {
		return k.Clone()
	}

	underlying := make(map[storage.UpgradeProgress_UpgradeState]struct{}, len(k)+len(other))
	for elem := range k {
		underlying[elem] = struct{}{}
	}
	for elem := range other {
		underlying[elem] = struct{}{}
	}
	return underlying
}

// Equal returns a bool if the sets are equal
func (k StorageUpgradeProgress_UpgradeStateSet) Equal(other StorageUpgradeProgress_UpgradeStateSet) bool {
	thisL, otherL := len(k), len(other)
	if thisL == 0 && otherL == 0 {
		return true
	}
	if thisL != otherL {
		return false
	}
	for elem := range k {
		if _, ok := other[elem]; !ok {
			return false
		}
	}
	return true
}

// AsSlice returns a slice of the elements in the set. The order is unspecified.
func (k StorageUpgradeProgress_UpgradeStateSet) AsSlice() []storage.UpgradeProgress_UpgradeState {
	if len(k) == 0 {
		return nil
	}
	elems := make([]storage.UpgradeProgress_UpgradeState, 0, len(k))
	for elem := range k {
		elems = append(elems, elem)
	}
	return elems
}

// GetArbitraryElem returns an arbitrary element from the set.
// This can be useful if, for example, you know the set has exactly one
// element, and you want to pull it out.
// If the set is empty, the zero value is returned.
func (k StorageUpgradeProgress_UpgradeStateSet) GetArbitraryElem() (arbitraryElem storage.UpgradeProgress_UpgradeState) {
	for elem := range k {
		arbitraryElem = elem
		break
	}
	return arbitraryElem
}

// AsSortedSlice returns a slice of the elements in the set, sorted using the passed less function.
func (k StorageUpgradeProgress_UpgradeStateSet) AsSortedSlice(less func(i, j storage.UpgradeProgress_UpgradeState) bool) []storage.UpgradeProgress_UpgradeState {
	slice := k.AsSlice()
	if len(slice) < 2 {
		return slice
	}
	// Since we're generating the code, we might as well use sort.Sort
	// and avoid paying the reflection penalty of sort.Slice.
	sortable := &sortableStorageUpgradeProgress_UpgradeStateSlice{slice: slice, less: less}
	sort.Sort(sortable)
	return sortable.slice
}

// Clear empties the set
func (k *StorageUpgradeProgress_UpgradeStateSet) Clear() {
	*k = nil
}

// Freeze returns a new, frozen version of the set.
func (k StorageUpgradeProgress_UpgradeStateSet) Freeze() FrozenStorageUpgradeProgress_UpgradeStateSet {
	return NewFrozenStorageUpgradeProgress_UpgradeStateSetFromMap(k)
}

// ElementsString returns a string representation of all elements, with individual element strings separated by `sep`.
// The string representation of an individual element is obtained via `fmt.Fprint`.
func (k StorageUpgradeProgress_UpgradeStateSet) ElementsString(sep string) string {
	if len(k) == 0 {
		return ""
	}
	var sb strings.Builder
	first := true
	for elem := range k {
		if !first {
			sb.WriteString(sep)
		}
		fmt.Fprint(&sb, elem)
		first = false
	}
	return sb.String()
}

// NewStorageUpgradeProgress_UpgradeStateSet returns a new thread unsafe set with the given key type.
func NewStorageUpgradeProgress_UpgradeStateSet(initial ...storage.UpgradeProgress_UpgradeState) StorageUpgradeProgress_UpgradeStateSet {
	underlying := make(map[storage.UpgradeProgress_UpgradeState]struct{}, len(initial))
	for _, elem := range initial {
		underlying[elem] = struct{}{}
	}
	return underlying
}

type sortableStorageUpgradeProgress_UpgradeStateSlice struct {
	slice []storage.UpgradeProgress_UpgradeState
	less  func(i, j storage.UpgradeProgress_UpgradeState) bool
}

func (s *sortableStorageUpgradeProgress_UpgradeStateSlice) Len() int {
	return len(s.slice)
}

func (s *sortableStorageUpgradeProgress_UpgradeStateSlice) Less(i, j int) bool {
	return s.less(s.slice[i], s.slice[j])
}

func (s *sortableStorageUpgradeProgress_UpgradeStateSlice) Swap(i, j int) {
	s.slice[j], s.slice[i] = s.slice[i], s.slice[j]
}

// A FrozenStorageUpgradeProgress_UpgradeStateSet is a frozen set of storage.UpgradeProgress_UpgradeState elements, which
// cannot be modified after creation. This allows users to use it as if it were
// a "const" data structure, and also makes it slightly more optimal since
// we don't have to lock accesses to it.
type FrozenStorageUpgradeProgress_UpgradeStateSet struct {
	underlying map[storage.UpgradeProgress_UpgradeState]struct{}
}

// NewFrozenStorageUpgradeProgress_UpgradeStateSetFromMap returns a new frozen set from the set-style map.
func NewFrozenStorageUpgradeProgress_UpgradeStateSetFromMap(m map[storage.UpgradeProgress_UpgradeState]struct{}) FrozenStorageUpgradeProgress_UpgradeStateSet {
	if len(m) == 0 {
		return FrozenStorageUpgradeProgress_UpgradeStateSet{}
	}
	underlying := make(map[storage.UpgradeProgress_UpgradeState]struct{}, len(m))
	for elem := range m {
		underlying[elem] = struct{}{}
	}
	return FrozenStorageUpgradeProgress_UpgradeStateSet{
		underlying: underlying,
	}
}

// NewFrozenStorageUpgradeProgress_UpgradeStateSet returns a new frozen set with the provided elements.
func NewFrozenStorageUpgradeProgress_UpgradeStateSet(elements ...storage.UpgradeProgress_UpgradeState) FrozenStorageUpgradeProgress_UpgradeStateSet {
	underlying := make(map[storage.UpgradeProgress_UpgradeState]struct{}, len(elements))
	for _, elem := range elements {
		underlying[elem] = struct{}{}
	}
	return FrozenStorageUpgradeProgress_UpgradeStateSet{
		underlying: underlying,
	}
}

// Contains returns whether the set contains the element.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) Contains(elem storage.UpgradeProgress_UpgradeState) bool {
	_, ok := k.underlying[elem]
	return ok
}

// Cardinality returns the cardinality of the set.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) Cardinality() int {
	return len(k.underlying)
}

// IsEmpty returns whether the underlying set is empty (includes uninitialized).
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) IsEmpty() bool {
	return len(k.underlying) == 0
}

// AsSlice returns the elements of the set. The order is unspecified.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) AsSlice() []storage.UpgradeProgress_UpgradeState {
	if len(k.underlying) == 0 {
		return nil
	}
	slice := make([]storage.UpgradeProgress_UpgradeState, 0, len(k.underlying))
	for elem := range k.underlying {
		slice = append(slice, elem)
	}
	return slice
}

// AsSortedSlice returns the elements of the set as a sorted slice.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) AsSortedSlice(less func(i, j storage.UpgradeProgress_UpgradeState) bool) []storage.UpgradeProgress_UpgradeState {
	slice := k.AsSlice()
	if len(slice) < 2 {
		return slice
	}
	// Since we're generating the code, we might as well use sort.Sort
	// and avoid paying the reflection penalty of sort.Slice.
	sortable := &sortableStorageUpgradeProgress_UpgradeStateSlice{slice: slice, less: less}
	sort.Sort(sortable)
	return sortable.slice
}

// ElementsString returns a string representation of all elements, with individual element strings separated by `sep`.
// The string representation of an individual element is obtained via `fmt.Fprint`.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) ElementsString(sep string) string {
	if len(k.underlying) == 0 {
		return ""
	}
	var sb strings.Builder
	first := true
	for elem := range k.underlying {
		if !first {
			sb.WriteString(sep)
		}
		fmt.Fprint(&sb, elem)
		first = false
	}
	return sb.String()
}

// The following functions make use of casting `k.underlying` into a mutable Set. This is safe, since we never leak
// references to these objects, and only invoke mutable set methods that are guaranteed to return a new copy.

// Union returns a frozen set that represents the union between this and other.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) Union(other FrozenStorageUpgradeProgress_UpgradeStateSet) FrozenStorageUpgradeProgress_UpgradeStateSet {
	if len(k.underlying) == 0 {
		return other
	}
	if len(other.underlying) == 0 {
		return k
	}
	return FrozenStorageUpgradeProgress_UpgradeStateSet{
		underlying: StorageUpgradeProgress_UpgradeStateSet(k.underlying).Union(other.underlying),
	}
}

// Intersect returns a frozen set that represents the intersection between this and other.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) Intersect(other FrozenStorageUpgradeProgress_UpgradeStateSet) FrozenStorageUpgradeProgress_UpgradeStateSet {
	return FrozenStorageUpgradeProgress_UpgradeStateSet{
		underlying: StorageUpgradeProgress_UpgradeStateSet(k.underlying).Intersect(other.underlying),
	}
}

// Difference returns a frozen set that represents the set difference between this and other.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) Difference(other FrozenStorageUpgradeProgress_UpgradeStateSet) FrozenStorageUpgradeProgress_UpgradeStateSet {
	return FrozenStorageUpgradeProgress_UpgradeStateSet{
		underlying: StorageUpgradeProgress_UpgradeStateSet(k.underlying).Difference(other.underlying),
	}
}

// Unfreeze returns a mutable set with the same contents as this frozen set. This set will not be affected by any
// subsequent modifications to the returned set.
func (k FrozenStorageUpgradeProgress_UpgradeStateSet) Unfreeze() StorageUpgradeProgress_UpgradeStateSet {
	return StorageUpgradeProgress_UpgradeStateSet(k.underlying).Clone()
}
