package mrutils

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"

	"github.com/OneOfOne/xxhash"
)

//Md5 Quickly Calculate md5 hash
func Md5(value string) string {
	// h := md5.New()
	// io.WriteString(h, value)
	return fmt.Sprintf("%x", md5.Sum([]byte(value)))
}

//SHA1 Quickly Calculate md5 hash
func SHA1(value []byte) string {
	// h := md5.New()
	// io.WriteString(h, value)
	return fmt.Sprintf("%x", sha1.Sum(value))
}

//XXHash32 Quickly Calculate md5 hash
func XXHASH32(value string) string {
	// h := md5.New()
	// io.WriteString(h, value)
	return fmt.Sprintf("%x", xxhash.ChecksumString32(value))
}

//XXHash64 Quickly Calculate md5 hash
func XXHASH64(value string) string {
	// h := md5.New()
	// io.WriteString(h, value)
	return fmt.Sprintf("%x", xxhash.ChecksumString64(value))
}
