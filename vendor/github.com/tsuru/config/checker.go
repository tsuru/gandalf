// Copyright 2014 Globo.com. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

type Checker func() error

// Check a parsed config file.
func Check(checkers []Checker) error {
	for _, check := range checkers {
		err := check()
		if err != nil {
			return err
		}
	}
	return nil
}
