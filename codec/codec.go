// Copyright 2025 Company.info B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package codec provides functionality for encoding and decoding data.
package codec

// Type is a string type used for encoding and decoding data.
type Type string

// Encoder is an interface that defines the Encode method, which takes any value
// and returns the encoded bytes and an error, if any.
type Encoder interface {
	Encode(v any) ([]byte, error)
}

// Decoder is an interface that defines the Decode method, which takes a byte
// slice of data and a value pointer, and decodes the data into the value.
type Decoder interface {
	Decode(data []byte, v any) error
}
