package flimsy

import "github.com/cespare/xxhash/v2"

type any interface{}

type callback func() (any, error)
type errorFunction func(error error)
type rootFunction func(dispose func()) any

var SYMBOL_ERRORS = int64(xxhash.Sum64String("SYMBOL_ERRORS") & 0x7fffffffffffffff)
