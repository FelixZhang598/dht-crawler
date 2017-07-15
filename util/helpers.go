/*
 * @author Felix Zhang
 * @email Javagoshell@gmail.com
 * @datetime 2017-07-14 21:22
 */

package util

import (
	"bytes"
	"fmt"
	"math/rand"
)

func RandByteString(length int) string {

	bf := bytes.NewBuffer(nil)
	for i := 0; i < length; i++ {
		bf.WriteByte(byte(rand.Intn(256)))
	}
	return fmt.Sprintf("%s", bf.Bytes())
}
