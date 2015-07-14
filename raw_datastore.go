// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gae

import (
	"fmt"

	"golang.org/x/net/context"
)

/// Kinds + Keys

// DSKey is the equivalent of *datastore.Key from the original SDK, except that
// it can have multiple implementations. See helper.DSKey* methods for missing
// methods like DSKeyIncomplete (and some new ones like DSKeyValid).
type DSKey interface {
	Kind() string
	StringID() string
	IntID() int64
	Parent() DSKey
	AppID() string
	Namespace() string

	String() string
}

// DSKeyTok is a single token from a multi-part DSKey.
type DSKeyTok struct {
	Kind     string
	IntID    int64
	StringID string
}

// DSCursor wraps datastore.Cursor.
type DSCursor interface {
	fmt.Stringer
}

// DSIterator wraps datastore.Iterator.
type DSIterator interface {
	Cursor() (DSCursor, error)
	Next(dst interface{}) (DSKey, error)
}

// DSQuery wraps datastore.Query.
type DSQuery interface {
	Ancestor(ancestor DSKey) DSQuery
	Distinct() DSQuery
	End(c DSCursor) DSQuery
	EventualConsistency() DSQuery
	Filter(filterStr string, value interface{}) DSQuery
	KeysOnly() DSQuery
	Limit(limit int) DSQuery
	Offset(offset int) DSQuery
	Order(fieldName string) DSQuery
	Project(fieldNames ...string) DSQuery
	Start(c DSCursor) DSQuery
}

// CommonDatastore is the interface for the methods which are common between
// Datastore and RawDatastore.
type CommonDatastore interface {
	NewKey(kind, stringID string, intID int64, parent DSKey) DSKey
	DecodeKey(encoded string) (DSKey, error)

	NewQuery(kind string) DSQuery
	Run(q DSQuery) DSIterator
	GetAll(q DSQuery, dst interface{}) ([]DSKey, error)
	Count(q DSQuery) (int, error)

	RunInTransaction(f func(c context.Context) error, opts *DSTransactionOptions) error
}

// RawDatastore implements the datastore functionality as described by
// the raw appengine documentation. No key inference occurs, nor does any
// caching. See Datastore for a nicer interface.
type RawDatastore interface {
	CommonDatastore

	Put(key DSKey, src interface{}) (DSKey, error)
	Get(key DSKey, dst interface{}) error
	Delete(key DSKey) error

	// These allow you to read and write a multiple datastore objects in
	// a non-atomic batch.
	DeleteMulti(keys []DSKey) error
	GetMulti(keys []DSKey, dst interface{}) error
	PutMulti(keys []DSKey, src interface{}) ([]DSKey, error)
}

// RDSFactory is the function signature for factory methods compatible with
// SetRDSFactory.
type RDSFactory func(context.Context) RawDatastore

// GetRDS gets the RawDatastore implementation from context.
func GetRDS(c context.Context) RawDatastore {
	if f, ok := c.Value(rawDatastoreKey).(RDSFactory); ok && f != nil {
		return f(c)
	}
	return nil
}

// SetRDSFactory sets the function to produce Datastore instances, as returned by
// the GetRDS method.
func SetRDSFactory(c context.Context, rdsf RDSFactory) context.Context {
	return context.WithValue(c, rawDatastoreKey, rdsf)
}

// SetRDS sets the current Datastore object in the context. Useful for testing
// with a quick mock. This is just a shorthand SetDSFactory invocation to set
// a factory which always returns the same object.
func SetRDS(c context.Context, rds RawDatastore) context.Context {
	return SetRDSFactory(c, func(context.Context) RawDatastore { return rds })
}
