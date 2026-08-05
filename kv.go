package main

import (
	"bytes"
	"encoding/gob"

	"github.com/syumai/workers/cloudflare/kv"
)

func GetEncodedKV[dataT any](key string, namespace string, data *dataT) error {
	ns, err := kv.NewNamespace(namespace)
	if err != nil {
		return err
	}
	r, err := ns.GetReader(key, nil)
	if err != nil {
		return err
	}
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
