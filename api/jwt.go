package api

import (
	"io"

	. "github.com/kitsudotfun/kyuubi/api/defs"

	"github.com/syumai/workers/cloudflare/kv"
)

func MustGetJwtKey(id string) []byte {
	keys, err := kv.NewNamespace(JwtKeyNamespace)
	if err != nil {
		panic(err)
	}
	r, err := keys.GetReader(id, nil)
	if err != nil {
		panic(err)
	}
	key, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return key
}
