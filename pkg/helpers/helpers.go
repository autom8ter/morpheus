package helpers

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"github.com/palantir/stacktrace"
	"github.com/spf13/cast"
	"strings"
	"time"
)

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func Uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

func JSONString(obj interface{}) string {
	bits, _ := json.MarshalIndent(obj, "", "    ")
	return string(bits)
}

func JWTExpired(token string) (bool, int64, error) {
	split := strings.Split(token, ".")
	if len(split) != 3 {
		return false, 0, stacktrace.NewError("invalid jwt, missing 3 segments")
	}
	payload := split[1]
	jsonPayload, err := base64.RawStdEncoding.DecodeString(payload)
	if err != nil {
		return false, 0, stacktrace.Propagate(err, "")
	}
	claims := map[string]interface{}{}
	if err := json.Unmarshal([]byte(jsonPayload), &claims); err != nil {
		return false, 0, err
	}
	exp, ok := claims["exp"]
	if !ok {
		return false, 0, stacktrace.NewError("missing exp claim")
	}
	expUnix := cast.ToInt64(exp)
	if expUnix > time.Now().Unix() {
		return false, expUnix, nil
	}
	return true, expUnix, nil
}
