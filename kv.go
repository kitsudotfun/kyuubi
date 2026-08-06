package main

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/syumai/workers/cloudflare/kv"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

func GetEncodedKV[dataT any](key string, namespace string, data *dataT) (err error) {
	ns, err := kv.NewNamespace(namespace)
	if err != nil {
		return err
	}
	r, err := ns.GetReader(key, nil)
	if err != nil {
		return err
	}

	// HACK: non-existent key returns null and decoding it panics
	// just recover for now...
	defer func() {
		r := recover()
		if r != nil {
			err = ErrKeyNotFound
		}
	}()

	err = gob.NewDecoder(r).Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func PutEncodedKV[dataT any](key string, namespace string, data dataT, opts *kv.PutOptions) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}
	ns, err := kv.NewNamespace(namespace)
	if err != nil {
		return err
	}
	err = ns.PutReader(key, &buf, opts)
	if err != nil {
		return err
	}

	return nil
}
