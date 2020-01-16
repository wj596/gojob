/*
 * Copyright 2020-2021 the original author(https://github.com/wj596)
 *
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * </p>
 */
package byteutil

import (
	"bytes"
	"encoding/binary"
)

func Uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func Uint8ToBytes(u uint8) ([]byte, error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	err := binary.Write(bytesBuffer, binary.BigEndian, &u)
	if err != nil {
		return nil, err
	}
	return bytesBuffer.Bytes(), nil
}

func BytesToUint8(b []byte) (uint8, error) {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint8
	err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	if err != nil {
		return 0, err
	}
	return tmp, nil
}
