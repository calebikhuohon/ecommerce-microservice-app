// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Modifications made:
// Hardcode currency as "USD"
// Remove functions sorely created for testing purposes

package money

import (
	pb "do-tutorial/src/checkoutservice/genproto"
	"errors"
)

const (
	nanosMin = -999999999
	nanosMax = +999999999
	nanosMod = 1000000000
)

var (
	ErrInValidValue = errors.New("one of the specified money values is invalid")
)

func IsValid(m pb.Money) bool {
	return signMatches(m) && validNanos(m.GetNanos())
}

func signMatches(m pb.Money) bool {
	return m.GetNanos() == 0 || m.GetUnits() == 0 || (m.GetNanos() < 0) == (m.GetUnits() < 0)
}

func validNanos(nanos int32) bool {
	return nanosMin <= nanos && nanos <= nanosMax
}

func Must(val pb.Money, err error) pb.Money {
	if err != nil {
		panic(err)
	}
	return val
}

func Sum(l, r pb.Money) (pb.Money, error) {
	if !IsValid(l) || !IsValid(r) {
		return pb.Money{}, ErrInValidValue
	}

	units := l.GetUnits() + r.GetUnits()
	nanos := l.GetNanos() + r.GetNanos()

	if (units == 0 && nanos == 0) || (units > 0 && nanos >= 0) || (units < 0 && nanos <= 0) {
		//same sign <units, nanos>
		units += int64(nanos / nanosMod)
		nanos = nanos % nanosMod
	} else {
		//different sign, nanos guaranteed to not go over the limit
		if units > 0 {
			units--
			nanos += nanosMod
		} else {
			units++
			nanos -= nanosMod
		}
	}

	return pb.Money{
		Units:        units,
		Nanos:        nanos,
		CurrencyCode: "USD",
	}, nil
}
